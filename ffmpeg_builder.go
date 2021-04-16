package main

import "fmt"

const ffmpegPath = "/home/jun/Sandbox/FFmpeg/ffmpeg"

func buildFfmpegStdinFLV2RTP(videoPort int, audioPort int) (*ffmpegCommand, error) {
	if videoPort <= 0 && audioPort <= 0 {
		return nil, fmt.Errorf("invalid args, missing videoPort and audioPort")
	}
	args := []string{
		"-loglevel",
		"level+info",
		"-f",
		"live_flv",
		"-i",
		"pipe:0",
	}

	if videoPort > 0 {
		args = append(args,
			"-map",
			"0:v",
			"-vcodec",
			"libvpx",
			"-g",
			"10",
			"-error-resilient",
			"1",
			"-quality",
			"realtime",
			"-cpu-used",
			"4",
			"-f",
			"rtp",
			fmt.Sprintf("rtp://127.0.0.1:%d?rtcpport=%d", videoPort, videoPort),
		)
	}
	if audioPort > 0 {
		args = append(args,
			"-map",
			"0:a",
			"-acodec",
			"libopus",
			"-f",
			"rtp",
			fmt.Sprintf("rtp://127.0.0.1:%d?rtcpport=%d", audioPort, audioPort),
		)
	}

	return newFfmpegCommand(ffmpegPath, args)
}

func buildFfmpegRTSP2RTP(url string, videoPort int, audioPort int) (*ffmpegCommand, error) {
	if videoPort <= 0 && audioPort <= 0 {
		return nil, fmt.Errorf("invalid args, missing videoPort and audioPort")
	}
	args := []string{
		"-loglevel",
		"level+info",
		"-f",
		"rtsp",
		"-rtsp_transport",
		"tcp",
		"-i",
		url,
	}

	if videoPort > 0 {
		args = append(args,
			"-map",
			"0:v",
			"-vcodec",
			"libvpx",
			"-g",
			"10",
			"-error-resilient",
			"1",
			"-quality",
			"realtime",
			"-cpu-used",
			"4",
			"-f",
			"rtp",
			fmt.Sprintf("rtp://127.0.0.1:%d?rtcpport=%d", videoPort, videoPort),
		)
	}
	if audioPort > 0 {
		args = append(args,
			"-map",
			"0:a",
			"-acodec",
			"libopus",
			"-f",
			"rtp",
			fmt.Sprintf("rtp://127.0.0.1:%d?rtcpport=%d", audioPort, audioPort),
		)
	}

	return newFfmpegCommand(ffmpegPath, args)
}
