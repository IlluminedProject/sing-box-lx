package wireguard

import (
	"encoding/base64"
	"encoding/hex"
	"net/netip"
	"os"
	"slices"
	"strings"
	"sync"

	E "github.com/sagernet/sing/common/exceptions"
	"go4.org/netipx"
)

// lx:begin runtime-management
// Package-level per-endpoint locks keep this file self-contained: no upstream
// struct change is needed to serialise peer replacement against lifecycle.
var runtimePeerAccess sync.Map

func runtimePeerLock(e *Endpoint) *sync.Mutex {
	lock, _ := runtimePeerAccess.LoadOrStore(e, new(sync.Mutex))
	return lock.(*sync.Mutex)
}

// ReplacePeers applies only the peer delta. Unchanged peers remain alive.
func (e *Endpoint) ReplacePeers(peers []PeerOptions) error {
	lock := runtimePeerLock(e)
	lock.Lock()
	defer lock.Unlock()
	if e.device == nil || e.tunDevice == nil || e.suspended.Load() {
		return os.ErrClosed
	}
	candidate, allowedPrefixes, err := buildPeerSnapshot(peers)
	if err != nil {
		return err
	}
	config := buildPeerDelta(e.peers, candidate)
	if config != "" {
		if err = e.device.IpcSet(config); err != nil {
			return E.Cause(err, "update wireguard peers")
		}
	}
	e.peers = candidate
	e.allowedAddress = allowedPrefixes
	e.allowedIPs = e.device.AllowedIPs()
	return nil
}

func buildPeerDelta(current, candidate []peerConfig) string {
	currentByKey := make(map[string]peerConfig, len(current))
	for _, peer := range current {
		currentByKey[peer.publicKeyHex] = peer
	}
	nextByKey := make(map[string]peerConfig, len(candidate))
	for _, peer := range candidate {
		nextByKey[peer.publicKeyHex] = peer
	}
	var config strings.Builder
	for publicKey := range currentByKey {
		if _, exists := nextByKey[publicKey]; !exists {
			config.WriteString("\npublic_key=" + publicKey + "\nremove=true")
		}
	}
	for publicKey, peer := range nextByKey {
		previous, exists := currentByKey[publicKey]
		if exists && peerConfigEqual(previous, peer) {
			continue
		}
		config.WriteString(peer.GenerateIpcLines())
	}
	return config.String()
}

func peerConfigEqual(left, right peerConfig) bool {
	return left.destination == right.destination &&
		left.endpoint == right.endpoint &&
		left.publicKeyHex == right.publicKeyHex &&
		left.preSharedKeyHex == right.preSharedKeyHex &&
		left.keepalive == right.keepalive &&
		left.reserved == right.reserved &&
		slices.Equal(left.allowedIPs, right.allowedIPs)
}

func buildPeerSnapshot(peers []PeerOptions) ([]peerConfig, []netip.Prefix, error) {
	candidate := make([]peerConfig, 0, len(peers))
	var allowedPrefixBuilder netipx.IPSetBuilder
	seen := make(map[string]struct{}, len(peers))
	for peerIndex, rawPeer := range peers {
		keepalive, err := awgKeepaliveSpec(rawPeer.PersistentKeepaliveInterval)
		if err != nil {
			return nil, nil, E.Cause(err, "persistent_keepalive_interval for peer ", peerIndex)
		}
		peer := peerConfig{
			allowedIPs: rawPeer.AllowedIPs,
			keepalive:  keepalive,
		}
		if rawPeer.Endpoint.Addr.IsValid() {
			peer.endpoint = rawPeer.Endpoint.AddrPort()
		} else if rawPeer.Endpoint.IsDomain() {
			peer.destination = rawPeer.Endpoint
		}
		publicKeyBytes, err := base64.StdEncoding.DecodeString(rawPeer.PublicKey)
		if err != nil || len(publicKeyBytes) != 32 {
			return nil, nil, E.New("invalid public key for peer ", peerIndex)
		}
		peer.publicKeyHex = hex.EncodeToString(publicKeyBytes)
		if _, found := seen[peer.publicKeyHex]; found {
			return nil, nil, E.New("duplicate peer public key")
		}
		seen[peer.publicKeyHex] = struct{}{}
		if rawPeer.PreSharedKey != "" {
			preSharedKeyBytes, err := base64.StdEncoding.DecodeString(rawPeer.PreSharedKey)
			if err != nil || len(preSharedKeyBytes) != 32 {
				return nil, nil, E.New("invalid pre shared key for peer ", peerIndex)
			}
			peer.preSharedKeyHex = hex.EncodeToString(preSharedKeyBytes)
		}
		if len(rawPeer.AllowedIPs) == 0 {
			return nil, nil, E.New("missing allowed ips for peer ", peerIndex)
		}
		if len(rawPeer.Reserved) > 0 {
			if len(rawPeer.Reserved) != 3 {
				return nil, nil, E.New("invalid reserved value for peer ", peerIndex, ", required 3 bytes, got ", len(rawPeer.Reserved))
			}
			copy(peer.reserved[:], rawPeer.Reserved[:])
		}
		for _, prefix := range rawPeer.AllowedIPs {
			allowedPrefixBuilder.AddPrefix(prefix)
		}
		candidate = append(candidate, peer)
	}
	allowedIPSet, err := allowedPrefixBuilder.IPSet()
	if err != nil {
		return nil, nil, err
	}
	return candidate, allowedIPSet.Prefixes(), nil
}

// lx:end runtime-management
