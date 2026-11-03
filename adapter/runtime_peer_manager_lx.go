package adapter

import "github.com/sagernet/sing-box/option"

// RuntimePeerManager replaces the peer set on a running WireGuard endpoint.
type RuntimePeerManager interface {
	ReplacePeers([]option.WireGuardPeer) error
	RuntimePeers() []option.WireGuardPeer
}
