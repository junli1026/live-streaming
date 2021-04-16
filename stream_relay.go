package main

import (
	"github.com/google/uuid"
	rtmp "github.com/junli1026/gortmp"
	"github.com/pion/webrtc/v2"
)

type streamRelay struct {
	id        string
	meta      *streamInfo
	ffmpeg    *ffmpegCommand
	peerGroup *PeerGroup
	errch     chan error
}

func newStreamRelay(id string) (*streamRelay, error) {
	var g *PeerGroup
	var err error
	g, err = newPeerGroup()
	if err != nil {
		return nil, err
	}
	logger.Infoln("create new node, video port %v, audio port %v", g.videoRTPPort, g.audioRTPPort)

	ffmpeg, err := buildFfmpegStdinFLV2RTP(g.videoRTPPort, g.audioRTPPort)
	if err != nil {
		g.stop()
		return nil, err
	}

	return &streamRelay{
		id:        id,
		peerGroup: g,
		errch:     make(chan error),
		ffmpeg:    ffmpeg,
	}, nil
}

func (relay *streamRelay) setStreamInfo(meta *rtmp.StreamMeta) {
	relay.meta = newStreamInfo(relay.id, meta)
}

func (relay *streamRelay) receiveUpStreamData(data []byte) error {
	return relay.ffmpeg.writeToStdin(data)
}

func (relay *streamRelay) acceptDownStreamClient(offer *webrtc.SessionDescription, streamName string) (string, *webrtc.SessionDescription, error) {
	id := uuid.New()
	answer, err := relay.peerGroup.acceptOffer(id, offer)
	if err != nil {
		return "", nil, err
	}
	return id.String(), answer, nil
}

func (relay *streamRelay) run() (err error) {
	go relay.peerGroup.run()
	return relay.ffmpeg.run()
}

func (relay *streamRelay) stop() {
	relay.ffmpeg.stop()
	relay.peerGroup.stop()
}
