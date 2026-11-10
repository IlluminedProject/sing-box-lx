package tls

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/logger"
	"github.com/stretchr/testify/require"
)

func makeInboundCertificate(t *testing.T, domain string) option.InboundTLSCertificateOptions {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: domain},
		DNSNames:     []string{domain},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	privateKey, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return option.InboundTLSCertificateOptions{
		Certificate: []string{string(certificate)},
		Key:         []string{string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKey}))},
	}
}

func TestInboundCertificatesSelectSNIAndFirstEntryFallback(t *testing.T) {
	fallback := makeInboundCertificate(t, "node.example.test")
	other := makeInboundCertificate(t, "global.example.test")
	config, err := NewSTDServer(context.Background(), logger.NOP(), option.InboundTLSOptions{
		Enabled:      true,
		Certificates: []option.InboundTLSCertificateOptions{fallback, other},
	})
	require.NoError(t, err)
	defer config.Close()
	serverConfig, err := config.STDConfig()
	require.NoError(t, err)

	for _, test := range []struct {
		name string
		sni  string
		want string
	}{
		{name: "known SNI", sni: "global.example.test", want: "global.example.test"},
		{name: "unknown SNI", sni: "unknown.example.test", want: "node.example.test"},
		{name: "no SNI", want: "node.example.test"},
	} {
		t.Run(test.name, func(t *testing.T) {
			clientConn, serverConn := net.Pipe()
			defer clientConn.Close()
			defer serverConn.Close()
			go func() {
				server := tls.Server(serverConn, serverConfig)
				_ = server.Handshake()
			}()
			client := tls.Client(clientConn, &tls.Config{
				ServerName:         test.sni,
				InsecureSkipVerify: true,
			})
			require.NoError(t, client.Handshake())
			require.Equal(t, test.want, client.ConnectionState().PeerCertificates[0].Subject.CommonName)
		})
	}
}
func TestInboundCertificateOptionsRejectMixedAndEmptySources(t *testing.T) {
	_, _, _, err := loadInboundCertificateOptions(context.Background(), nil)
	require.Error(t, err)

	_, _, _, err = loadInboundCertificateOptions(context.Background(), []option.InboundTLSCertificateOptions{{
		Certificate:     []string{"certificate"},
		CertificatePath: "certificate.pem",
		Key:             []string{"key"},
	}})
	require.Error(t, err)
}
func TestInboundCertificateOptionsRejectIncompletePair(t *testing.T) {
	_, _, _, err := loadInboundCertificateOptions(context.Background(), []option.InboundTLSCertificateOptions{{
		Certificate: []string{"certificate"},
	}})
	require.Error(t, err)
}
