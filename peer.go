package main

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

type Peer struct {
	id              uuid.UUID
	offer           webrtc.SessionDescription
	peerConn        *webrtc.PeerConnection
	videoTrack      *webrtc.Track
	audioTrack      *webrtc.Track
	videoRTPPayload uint8
	videoSSRC       uint32
	audioRTPPayload uint8
	audioSSRC       uint32
	vp8Supported    bool
	opusSupported   bool
}

func (p *Peer) writeVideoPacket(rtp *rtp.Packet) error {
	if p.videoTrack == nil {
		return errors.New("video track is empty")
	}
	// overwrite the payload type field in RTP packet
	rtp.Header.PayloadType = p.videoRTPPayload
	rtp.Header.SSRC = p.videoSSRC
	return p.videoTrack.WriteRTP(rtp)
}

func (p *Peer) writeAudioPacket(rtp *rtp.Packet) error {
	if p.audioTrack == nil {
		return errors.New("audio track is empty")
	}
	// overwrite the payload type field in RTP packet
	rtp.Header.PayloadType = p.audioRTPPayload
	rtp.Header.SSRC = p.audioSSRC
	return p.audioTrack.WriteRTP(rtp)
}

func (p *Peer) addVideoTrack() error {
	if p.videoTrack != nil {
		return errors.New("already has video track")
	}
	if !p.vp8Supported {
		return errors.New("the peer doesn't support vp8")
	}

	p.videoSSRC = rand.Uint32()
	video, err := p.peerConn.NewTrack(p.videoRTPPayload, p.videoSSRC, "video", "pion")
	if err != nil {
		return err
	}
	if _, err = p.peerConn.AddTrack(video); err != nil {
		return err
	}
	p.videoTrack = video
	return nil
}

func (p *Peer) addAudioTrack() error {
	if p.audioTrack != nil {
		return errors.New("already has audio track")
	}
	if !p.opusSupported {
		return errors.New("the peer doesn't support opus")
	}

	p.audioSSRC = rand.Uint32()
	audio, err := p.peerConn.NewTrack(p.audioRTPPayload, p.audioSSRC, "audio", "pion")
	if err != nil {
		return err
	}
	if _, err = p.peerConn.AddTrack(audio); err != nil {
		return err
	}
	p.audioTrack = audio
	return nil
}

func findCodecByName(engine *webrtc.MediaEngine, codecType webrtc.RTPCodecType, name string) *webrtc.RTPCodec {
	for _, codec := range engine.GetCodecsByKind(codecType) {
		if strings.ToLower(codec.Name) == strings.ToLower(name) {
			return codec
		}
	}
	return nil
}

func newPeer(id uuid.UUID, offer *webrtc.SessionDescription) (peer *Peer, err error) {
	// populate media engine from sdp
	mediaEngine := webrtc.MediaEngine{}
	if err = mediaEngine.PopulateFromSDP(*offer); err != nil {
		return nil, err
	}

	peer = &Peer{
		id:    id,
		offer: *offer,
	}

	videoCodec := findCodecByName(&mediaEngine, webrtc.RTPCodecTypeVideo, "VP8")
	if videoCodec != nil {
		peer.vp8Supported = true
		peer.videoRTPPayload = videoCodec.PayloadType
	} else {
		return nil, errors.New("vp8 codec support missing")
	}

	audioCodec := findCodecByName(&mediaEngine, webrtc.RTPCodecTypeAudio, "OPUS")
	if audioCodec != nil {
		peer.opusSupported = true
		peer.audioRTPPayload = audioCodec.PayloadType
	}

	// Create a new RTCPeerConnection
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
	if peer.peerConn, err = api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}); err != nil {
		return nil, err
	}

	return peer, nil
}

func (p *Peer) prepareAnswer() (*context.Context, *webrtc.SessionDescription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// This will notify you when the peer has connected/disconnected
	p.peerConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		logger.Infof("ICE connection state changed to '%s'\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			cancel()
		}
	})

	// Set the remote SessionDescription
	if err := p.peerConn.SetRemoteDescription(p.offer); err != nil {
		return nil, nil, err
	}

	// Create answer
	answer, err := p.peerConn.CreateAnswer(nil)
	if err != nil {
		return nil, nil, err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err = p.peerConn.SetLocalDescription(answer); err != nil {
		return nil, nil, err
	}
	return &ctx, &answer, nil
}

func (p *Peer) close() error {
	if err := p.peerConn.Close(); err != nil {
		return err
	}
	return nil
}
