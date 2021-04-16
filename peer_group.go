package main

import (
	"context"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

const startPort = 9001
const endPort = 9001 + 1000

type PeerGroup struct {
	videoRTPPort int
	audioRTPPort int
	videoRTPConn *net.UDPConn
	audioRTPConn *net.UDPConn
	peers        []*Peer
	mutex        sync.Mutex
}

func findAvailableConn() (int, *net.UDPConn, error) {
	var port int
	var conn *net.UDPConn
	var err error
	for port = startPort; port <= endPort; port++ {
		conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
		if err == nil {
			return port, conn, nil
		}
	}
	return -1, nil, err
}

func newPeerGroup() (*PeerGroup, error) {
	videoPort, videoConn, err := findAvailableConn()
	if err != nil {
		return nil, err
	}

	audioPort, audioConn, err := findAvailableConn()
	if err != nil {
		return nil, err
	}

	return &PeerGroup{
		videoRTPPort: videoPort,
		videoRTPConn: videoConn,
		audioRTPPort: audioPort,
		audioRTPConn: audioConn,
		peers:        make([]*Peer, 0),
	}, nil
}

func (g *PeerGroup) addPeer(p *Peer) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.peers = append(g.peers, p)
}

func (g *PeerGroup) acceptOffer(id uuid.UUID, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	peer, err := newPeer(id, offer)
	if err != nil {
		return nil, err
	}
	peer.addVideoTrack()
	peer.addAudioTrack()

	iceConnected, sdp, err := peer.prepareAnswer()
	if err != nil {
		return nil, err
	}
	go func() {
		<-(*iceConnected).Done()
		err := (*iceConnected).Err()
		if err != context.Canceled {
			logger.Errorf("failed to build ice connection with peer %v,", peer.id)
			return
		}
		g.addPeer(peer)
	}()
	return sdp, nil
}

func (g *PeerGroup) receiveVideoPacket(pkt *rtp.Packet) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for _, peer := range g.peers {
		if !peer.vp8Supported {
			continue
		}
		if err := peer.writeVideoPacket(pkt); err != nil {
			logger.Errorf("failed to write video packet to peer %v, %v\n", peer.id, err)
		}
	}
}

func (g *PeerGroup) receiveAudioPacket(pkt *rtp.Packet) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for _, peer := range g.peers {
		if !peer.opusSupported {
			continue
		}
		if err := peer.writeAudioPacket(pkt); err != nil {
			logger.Errorf("failed to write audio packet to peer %v, %v\n", peer.id, err)
		}
	}
}

func forwardPacket(conn *net.UDPConn, receive func(*rtp.Packet)) error {
	buf := make([]byte, 4096) // UDP MTU
	// Read RTP package and dispatch to all peers
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			logger.Errorf("failed to read from  %v, %v\n", conn.LocalAddr, err)
			return err
		}

		packet := &rtp.Packet{}
		if err := packet.Unmarshal(buf[:n]); err != nil {
			logger.Errorln(err)
			return err
		}

		receive(packet)
	}
}

func (g *PeerGroup) run() error {
	ch := make(chan error)
	go func() {
		ch <- forwardPacket(g.videoRTPConn, g.receiveVideoPacket)
	}()
	go func() {
		ch <- forwardPacket(g.audioRTPConn, g.receiveAudioPacket)
	}()
	return <-ch
}

func (g *PeerGroup) stop() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	for _, peer := range g.peers {
		if err := peer.close(); err != nil {
			logger.Errorf("failed to close peer %v\n", peer.id)
		} else {
			logger.Infof("close peer %v done.\n", peer.id)
		}
	}
}
