package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents a remote node over a TCP established connection
type TCPPeer struct {
	// conn is the underlying connection of the peer
	conn net.Conn

	// outbound = true if we dial and retrieve a connection,
	// and it's false if we accept and retrieve a connection
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddr  string
	HandshakeFn HandshakeFunc
	Decoder     Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func (trt *TCPTransport) ListenAndAccept() error {
	var err error
	trt.listener, err = net.Listen("tcp", trt.ListenAddr)
	if err != nil {
		return err
	}
	go trt.startAcceptLoop()
	return nil
}

func (trt *TCPTransport) startAcceptLoop() {
	for {
		conn, err := trt.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		fmt.Printf("new incoming connection: %+v\n", conn)

		go trt.handleConn(conn)
	}
}

func (trt *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := trt.HandshakeFn(peer); err != nil {
		fmt.Printf("TCP handshake error: %s\n", err)
		conn.Close()
		return
	}
	msg := &Message{}
	for {
		if err := trt.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
		msg.From = conn.RemoteAddr()
		fmt.Printf("message: %+v\n", msg)
	}
}
