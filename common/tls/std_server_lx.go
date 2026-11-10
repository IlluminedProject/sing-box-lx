package tls

import (
	"context"
	stdtls "crypto/tls"
	"strings"

	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/service/filemanager"
)

func loadInboundCertificateOptions(ctx context.Context, options []option.InboundTLSCertificateOptions) ([]stdtls.Certificate, []string, []string, error) {
	if len(options) == 0 {
		return nil, nil, nil, E.New("missing certificates")
	}
	certificates := make([]stdtls.Certificate, len(options))
	certificatePaths := make([]string, len(options))
	keyPaths := make([]string, len(options))
	for index, option := range options {
		inlineCertificate := len(option.Certificate) > 0
		inlineKey := len(option.Key) > 0
		pathCertificate := option.CertificatePath != ""
		pathKey := option.KeyPath != ""
		if inlineCertificate == pathCertificate || inlineKey == pathKey || inlineCertificate != inlineKey || pathCertificate != pathKey {
			return nil, nil, nil, E.New("certificate entry ", index, " must contain one matching inline or path certificate/key pair")
		}
		var certificatePEM, keyPEM []byte
		var err error
		if inlineCertificate {
			certificatePEM = []byte(strings.Join(option.Certificate, "\n"))
			keyPEM = []byte(strings.Join(option.Key, "\n"))
		} else {
			certificatePaths[index] = option.CertificatePath
			keyPaths[index] = option.KeyPath
			certificatePEM, err = filemanager.ReadFile(ctx, option.CertificatePath)
			if err != nil {
				return nil, nil, nil, E.Cause(err, "read certificate ", index)
			}
			keyPEM, err = filemanager.ReadFile(ctx, option.KeyPath)
			if err != nil {
				return nil, nil, nil, E.Cause(err, "read key ", index)
			}
		}
		certificates[index], err = stdtls.X509KeyPair(certificatePEM, keyPEM)
		if err != nil {
			return nil, nil, nil, E.Cause(err, "parse certificate entry ", index)
		}
	}
	return certificates, certificatePaths, keyPaths, nil
}

func loadInboundCertificates(ctx context.Context, options []option.InboundTLSCertificateOptions) ([]stdtls.Certificate, error) {
	certificates, _, _, err := loadInboundCertificateOptions(ctx, options)
	return certificates, err
}
