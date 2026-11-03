package wireguard

import (
	"net/netip"
	"os"
	"sync"
	"sync/atomic"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	transport "github.com/sagernet/sing-box/transport/wireguard"
	M "github.com/sagernet/sing/common/metadata"
)

var _ adapter.RuntimePeerManager = (*Endpoint)(nil)

type runtimePeerState struct {
	access sync.Mutex
	peers  atomic.Pointer[[]option.WireGuardPeer]
}

var runtimePeerStates sync.Map

func (w *Endpoint) peerState() *runtimePeerState {
	state, _ := runtimePeerStates.LoadOrStore(w, &runtimePeerState{})
	return state.(*runtimePeerState)
}

func (w *Endpoint) ReplacePeers(peers []option.WireGuardPeer) error {
	state := w.peerState()
	state.access.Lock()
	defer state.access.Unlock()
	if w.closing.Load() || !w.started.Load() || w.idleAsleep.Load() || w.torndown.Load() || w.building.Load() {
		return os.ErrClosed
	}
	options := make([]transport.PeerOptions, len(peers))
	for i, peer := range peers {
		if len(peer.AllowedIPs) == 0 || peer.PublicKey == "" {
			return os.ErrInvalid
		}
		options[i] = transport.PeerOptions{
			Endpoint:                    M.ParseSocksaddrHostPort(peer.Address, peer.Port),
			PublicKey:                   peer.PublicKey,
			PreSharedKey:                peer.PreSharedKey,
			AllowedIPs:                  append([]netip.Prefix(nil), peer.AllowedIPs...),
			PersistentKeepaliveInterval: peer.PersistentKeepaliveInterval,
			Reserved:                    append([]uint8(nil), peer.Reserved...),
		}
	}
	if err := w.endpoint.ReplacePeers(options); err != nil {
		return err
	}
	copyPeers := append([]option.WireGuardPeer(nil), peers...)
	state.peers.Store(&copyPeers)
	return nil
}

func (w *Endpoint) RuntimePeers() []option.WireGuardPeer {
	peers := w.peerState().peers.Load()
	if peers == nil {
		return nil
	}
	result := make([]option.WireGuardPeer, len(*peers))
	for i, peer := range *peers {
		peer.PreSharedKey = ""
		peer.AllowedIPs = append([]netip.Prefix(nil), peer.AllowedIPs...)
		peer.Reserved = append([]uint8(nil), peer.Reserved...)
		result[i] = peer
	}
	return result
}
