package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
)

type offerMessage struct {
	Stream string                    `json: "stream"`
	Sdp    webrtc.SessionDescription `json: "sdp"`
}

type answerMessage struct {
	PeerID string                    `json:"peerId"`
	Sdp    webrtc.SessionDescription `json:"sdp"`
}

type signalHandler struct {
	relaies  relayAccessor
	upgrader websocket.Upgrader
}

func newSignalHandler(accessor relayAccessor) *signalHandler {
	return &signalHandler{
		relaies: accessor,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (s *signalHandler) handle(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		conn.WriteJSON(gin.H{
			"error": "Failed to set websocket upgrade",
		})
		return
	}
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		offer := offerMessage{}
		err = json.Unmarshal(msg, &offer)
		if err != nil {
			conn.WriteJSON(gin.H{
				"error": err,
			})
			return
		}

		relay := s.relaies.getRelay(offer.Stream)
		if relay == nil {
			conn.WriteJSON(gin.H{
				"error": fmt.Sprintf("stream %v not found", offer.Stream),
			})
			return
		}
		peerID, answer, err := relay.acceptDownStreamClient(&offer.Sdp, offer.Stream)
		if err != nil {
			conn.WriteJSON(gin.H{
				"error": err,
			})
			return
		}
		answerMessage := answerMessage{
			PeerID: peerID,
			Sdp:    *answer,
		}
		conn.WriteJSON(answerMessage)
	}
}
