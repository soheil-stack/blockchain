package peer

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type Peer struct {
	Host string
}

func New(Host string) Peer {
	return Peer{
		Host: Host,
	}
}

func (peer Peer) Match(host string) bool {
	return peer.Host == host
}

type PeerSet struct {
	mu  sync.RWMutex
	set map[Peer]struct{}
}

func NewPeerSet() *PeerSet {
	return &PeerSet{
		set: make(map[Peer]struct{}),
	}
}

func (ps *PeerSet) Add(peer Peer) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, exists := ps.set[peer]
	if !exists {
		ps.set[peer] = struct{}{}
		return true
	}

	return false
}

func (ps *PeerSet) Remove(peer Peer) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	delete(ps.set, peer)
}

func (ps *PeerSet) Copy(host string) []Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var peers []Peer
	for peer := range ps.set {
		if !peer.Match(host) {
			peers = append(peers, peer)
		}
	}

	return peers
}

type PeerStatus struct {
	LatestBlockHash   common.Hash `json:"latestBlockHash"`
	LatestBlockNumber uint64      `json:"latestBlockNumber"`
	KnownPeers        []Peer      `json:"knownPeers"`
}
