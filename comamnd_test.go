package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

func printLine(r io.ReadCloser) error {
	reader := bufio.NewReader(r)
	defer r.Close()
	for {
		l, _, err := reader.ReadLine()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Println(string(l))
	}
}

func TestFfmpeg(t *testing.T) {
	args := []string{
		"-loglevel",
		"trace",
		"-hide_banner",
		"-f",
		"lavfi",
		"-i",
		"testsrc=duration=100:size=1280x720:rate=30",
		"-vcodec",
		"libvp8",
		"-f",
		"mp4",
		"-movflags",
		"frag_keyframe+empty_moov",
		"pipe:1",
	}
	r, _ := newFfmpegCommand("/home/jun/Sandbox/FFmpeg/ffmpeg", args)

	ch := make(chan int, 2)
	go func() {
		reader := bufio.NewReader(r.stdout)
		f, err := os.Create("./test.mp4")
		buf := make([]byte, 1024)
		defer f.Close()
		defer r.stdout.Close()
		var n int
		for {
			n, err = reader.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fail()
			}
			c := 0
			i := 0
			for c < n {
				i, err = f.Write(buf[c:n])
				if err != nil {
					t.Fail()
				}
				c += i
			}
		}
		ch <- 1
	}()

	go func() {
		printLine(r.stderr)
		ch <- 1
	}()

	go r.run()

	time.Sleep(2 * time.Second)
	r.stop()
	<-ch
	<-ch
}

func TestFfmpeg2(t *testing.T) {
	args := []string{
		"-loglevel",
		"trace",
		"-hide_banner",
		"-f",
		"live_flv",
		"-i",
		"-",
		"-vcodec",
		"libvp8",
		"-f",
		"mp4",
		"-movflags",
		"frag_keyframe+empty_moov",
		"pipe:1",
	}
	r, _ := newFfmpegCommand("/home/jun/Sandbox/FFmpeg/ffmpeg", args)

	ch := make(chan int, 2)
	go func() {
		reader := bufio.NewReader(r.stdout)
		f, err := os.Create("./test.mp4")
		buf := make([]byte, 1024)
		defer f.Close()
		defer r.stdout.Close()
		var n int
		for {
			n, err = reader.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fail()
			}
			c := 0
			i := 0
			for c < n {
				i, err = f.Write(buf[c:n])
				if err != nil {
					t.Fail()
				}
				c += i
			}
		}
		ch <- 1
	}()

	go func() {
		printLine(r.stderr)
		ch <- 1
	}()

	go r.run()

	time.Sleep(2 * time.Second)
	r.stop()
	<-ch
	<-ch
}
