package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type httpService struct {
	*gin.Engine
	relaies   relayAccessor
	registrar streamRegistrar
	signal    *signalHandler
}

func newHTTPService(accessor relayAccessor, registrar streamRegistrar) *httpService {
	r := gin.Default()
	hs := &httpService{
		r,
		accessor,
		registrar,
		newSignalHandler(accessor),
	}
	r.Use(apiMiddleware(hs))
	r.Static("/public", "./public")
	api := r.Group("/api")
	{
		//api.GET("/streams", listStreams)
		//api.POST("/streams", createStream)
		api.GET("/ws", func(c *gin.Context) {
			s, ok := c.MustGet("server").(*httpService)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}
			s.signal.handle(c.Writer, c.Request)
		})

	}
	return hs
}

func apiMiddleware(hs *httpService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("server", hs)
		c.Next()
	}
}

/*
type ClientSDP struct {
	Sdp string `json: "sdp"`
}

func receiveSDP(c *gin.Context) {
	_, ok := c.MustGet("server").(*server)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var clientSDP ClientSDP
	err := c.BindJSON(&clientSDP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	sdp := webrtc.SessionDescription{}
	b, err := base64.StdEncoding.DecodeString(clientSDP.Sdp)
	if err != nil {
		logger.Errorf("failed to decode base64 client sdp %v\n", err)
	}
	err = json.Unmarshal(b, &sdp)
	if err != nil {
		logger.Errorf("failed to unmarshal json of sdp %v\n", err)
	}
	c.JSON(http.StatusOK, gin.H{})
}

*/

func listStreams(c *gin.Context) {
	s, ok := c.MustGet("server").(*httpService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	streamIds := s.registrar.listStreamIds()
	c.JSON(http.StatusOK, gin.H{
		"streams": streamIds,
	})
}

type streamEntity struct {
	StreamName string `json:"streamName"`
	StreamID   string `json:"streamId"`
}

func createStream(c *gin.Context) {
	s, ok := c.MustGet("server").(*httpService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	var st streamEntity
	err := c.Bind(&st)
	if err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error(),
		})
	}
	if len(s.registrar.listStreamIds()) >= 2 {
		c.JSON(http.StatusLocked, gin.H{})
		return
	}
	st.StreamID = s.registrar.registerStream(st.StreamName)
	c.JSON(http.StatusCreated, st)
}
