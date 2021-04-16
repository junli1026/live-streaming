package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"

	rtmp "github.com/junli1026/gortmp"
)

func validateRTMPStreamURL(uri string) error {
	url, err := url.Parse(uri)
	if err != nil {
		return err
	}
	if url.Path != "/live" && url.Path != "/LIVE" {
		return errors.New("invalid string url " + uri)
	}
	return nil
}

func newRTMPService(rtmpAddr string, accessor relayAccessor, registrar streamRegistrar) *rtmp.RtmpServer {
	s := rtmp.NewServer()
	s.OnStreamData(func(meta *rtmp.StreamMeta, data *rtmp.StreamData) error {
		streamKey := meta.StreamName()
		if data.Type == rtmp.FlvHeader { //stream begin
			if err := validateRTMPStreamURL(meta.URL()); err != nil {
				return err
			}

			if _, ok := registrar.checkStream(streamKey); !ok {
				return fmt.Errorf("unauthenticated stream key %v", streamKey)
			}

			r := accessor.getRelay(streamKey)
			if r == nil {
				logger.Infof("relay runner for %s not exist, create one", streamKey)
				r = accessor.createRelay(streamKey)
				if r == nil {
					return fmt.Errorf("failed to create relay for stream %v", streamKey)
				}
				r.setStreamInfo(meta)
				go r.run()
			}
		}
		relay := accessor.getRelay(streamKey)
		return relay.receiveUpStreamData(data.Data)
	})

	s.OnStreamClose(func(meta *rtmp.StreamMeta, err error) {
		streamKey := meta.StreamName()
		msg := fmt.Sprintf("stream key %v stopped for reason: %v\n", streamKey, err)
		if err == io.EOF {
			logger.Info(msg)
		} else {
			logger.Error(msg)
		}
		relay := accessor.getRelay(streamKey)
		if relay == nil {
			logger.Errorf("relay with stream key %v not found.\n", streamKey)
			return
		}
		relay.stop()
		accessor.deleteRelay(streamKey)
	})

	return s
}
