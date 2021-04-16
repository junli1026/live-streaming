package main

import (
	rtmp "github.com/junli1026/gortmp"
)

type streamInfo struct {
	StreamKey       string `json:"streamKey"`
	Encoder         string `json:"encoder"`
	VideoCodec      string `json:"videoCodec"`
	VideoHeight     int    `json:"videoHeight"`
	VideoWidth      int    `json:"videoWidth"`
	VideoFrameRate  int    `json:"videoFrameRate"`
	AudioCodec      string `json:"audioCodec"`
	AudioChannels   int    `json:"audioChannels"`
	AudioSampleRate int    `json:"audioSampleRate"`
	AudioSampleSize int    `json:"audioSampleSize"`
}

func newStreamInfo(streamKey string, meta *rtmp.StreamMeta) *streamInfo {
	meta.VideoCodec()
	return &streamInfo{
		StreamKey:       streamKey,
		Encoder:         meta.Encoder(),
		VideoCodec:      meta.VideoCodec(),
		VideoHeight:     meta.Height(),
		VideoFrameRate:  meta.FrameRate(),
		VideoWidth:      meta.Width(),
		AudioCodec:      meta.AudioCodec(),
		AudioChannels:   meta.AudioChannels(),
		AudioSampleRate: meta.AudioSampleRate(),
		AudioSampleSize: meta.AudioSampleSize(),
	}
}
