package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

// TODO: use message types like ClientName, TextMessage, TokenVerification

const Port = ":3000"

type Server struct {
	Clients map[*websocket.Conn]struct{}
}

type MessageResponse struct {
	Sender string `json:"sender"`
	Text   string `json:"text"`
}
type MessageRequest struct {
	Text string `json:"text"`
}

func NewServer() *Server {
	return &Server{
		Clients: make(map[*websocket.Conn]struct{}),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	log.Printf("new connection from client: %s", ws.RemoteAddr())
	s.Clients[ws] = struct{}{}
	s.readLoop(ws)
}

func getClientName(ws *websocket.Conn) string {
	buf := make([]byte, 1024)
	n, err := ws.Read(buf)
	if err != nil {
		log.Printf("error reading client name (%s): %v", ws.RemoteAddr(), err)
		return "UNKNOWN"
	}

	msg := struct {
		Name string `json:"name"`
	}{}

	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		log.Printf("error unmarshal client name (%s): %v", ws.RemoteAddr(), err)
		return "UNKNOWN"
	}

	return msg.Name
}

func (s *Server) readLoop(ws *websocket.Conn) {
	defer func() {
		ws.Close()
		log.Printf("connection closed with client: %s", ws.RemoteAddr())
	}()

	name := getClientName(ws)

	for {
		buf := make([]byte, 1024)
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				delete(s.Clients, ws)
				return
			}
			log.Printf("error reading msg (%s): %v", ws.RemoteAddr(), err)
			continue
		}

		var msg MessageRequest
		err = json.Unmarshal(buf[:n], &msg)
		if err != nil {
			log.Printf("error unmarshal msg (%s): %v", ws.RemoteAddr(), err)
			continue
		}

		log.Printf("new message (%s): %v", ws.RemoteAddr(), msg.Text)

		s.broadcast(&MessageResponse{
			Text:   msg.Text,
			Sender: name,
		})
	}
}

func (s *Server) broadcast(msg *MessageResponse) {
	for ws := range s.Clients {
		go func(ws *websocket.Conn) {
			if err := websocket.JSON.Send(ws, msg); err != nil {
				log.Printf("error sending msg (%s): %v", ws.RemoteAddr(), err)
			}
		}(ws)
	}
}

func main() {
	s := NewServer()
	http.Handle("/ws", websocket.Handler(s.handleWS))

	log.Printf("starting server on port %v", Port)
	http.ListenAndServe(Port, nil)
}
