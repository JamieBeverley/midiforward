package server

import (
	"encoding/json"
	"fmt"
	"midiforward/internal/forwardport"
	"net"
	"time"
)

type Server struct {
	Addr        *net.UDPAddr
	ForwardPort *forwardport.ForwardPort
}

type TimedMidiMessage struct {
	Message []int
	When    int64
}

func (tmm TimedMidiMessage) toBytes() []byte {
	bytes := make([]byte, len(tmm.Message))
	for i, v := range tmm.Message {
		bytes[i] = byte(v)
	}
	return bytes
}

func parseTimedMidiMessage(msg []byte) (TimedMidiMessage, error) {
	var tmm TimedMidiMessage
	err := json.Unmarshal(msg, &tmm)
	return tmm, err
}

func forwardWhen(tmm TimedMidiMessage, forwardPort *forwardport.ForwardPort) {
	now := int64(time.Now().UnixMilli())
	sleepMs := tmm.When - now
	fmt.Printf("Sleeping for %d ms\n", sleepMs)
	if sleepMs > 0 {
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}
	forwardPort.ForwardMessage(tmm.toBytes(), 0)
}

func handleMessage(conn *net.UDPConn, forwardPort *forwardport.ForwardPort) {
	json := make([]byte, 1024)
	var err error
	n, _, err := conn.ReadFromUDP(json)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	tmm, err := parseTimedMidiMessage(json[:n])
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	go forwardWhen(tmm, forwardPort)
}

func New(addr string, port int, forwardPort *forwardport.ForwardPort) *Server {
	udpAddr := net.UDPAddr{Port: int(port), IP: net.ParseIP(addr)}
	return &Server{Addr: &udpAddr, ForwardPort: forwardPort}
}

func (server *Server) Listen() error {
	conn, err := net.ListenUDP("udp", server.Addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	for {
		handleMessage(conn, server.ForwardPort)
	}
}
