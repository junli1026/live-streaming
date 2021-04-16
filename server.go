package main

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	rtmp "github.com/junli1026/gortmp"
	"github.com/junli1026/gortmp/logging"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *logrus.Logger = logging.Logger

type server struct {
	rtmp      *rtmp.RtmpServer
	http      *httpService
	relaies   *relayManager
	registrar *streamRegistrarImpl
}

func initLogger() {
	l := lumberjack.Logger{
		//Filename:   "test.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     10,    // days
		Compress:   false, // disabled by default
	}

	out := io.MultiWriter(&l, os.Stderr)

	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(out)
	gin.DefaultWriter = out
	gin.DefaultErrorWriter = out
}

func newSever(rtmpAddr string) *server {
	initLogger()
	relaies := newRelayManager()
	registrar := newStreamRegistrar()
	rtmp := newRTMPService(rtmpAddr, relaies, registrar)
	http := newHTTPService(relaies, registrar)
	s := &server{
		rtmp:      rtmp,
		http:      http,
		relaies:   relaies,
		registrar: registrar,
	}

	return s
}

func (s *server) run() (err error) {
	c := make(chan error)
	go func() {
		c <- s.rtmp.Run(":1936")
	}()

	go func() {
		c <- s.http.Run(":9000")
	}()

	for i := 0; i < 2; i++ {
		select {
		case err := <-c:
			{
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
