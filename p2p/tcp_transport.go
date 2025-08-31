package p2p

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/rs/zerolog"
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

func (pr *TCPPeer) Close() error {
	return pr.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr  string
	HandshakeFn HandshakeFunc
	OnPeer      func(Peer) error
	Decoder     Decoder
}

type TCPTransport struct {
	opts     TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
	log      zerolog.Logger

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	writer := zerolog.NewConsoleWriter()
	logger := zerolog.New(writer).With().Timestamp().Logger()

	return &TCPTransport{opts: opts,
		rpcch: make(chan RPC),
		log:   logger,
	}
}

func (prt *TCPTransport) ListenAndAccept() error {
	var err error
	prt.listener, err = net.Listen("tcp", prt.opts.ListenAddr)
	if err != nil {
		return err
	}
	go prt.startAcceptLoop()
	prt.log.Info().Msgf("TCP transport listening on port=%s", prt.opts.ListenAddr)
	return nil
}

// Consume will return a read-only channel for reading the incoming messages
// received from another peer in the network.
func (prt *TCPTransport) Consume() <-chan RPC { return prt.rpcch }

func (prt *TCPTransport) Close() error { return prt.listener.Close() }

func (prt *TCPTransport) startAcceptLoop() {
	for {
		conn, err := prt.listener.Accept()
		if err != nil {
			prt.log.Err(err).Msg("TCP accept error")
		}
		prt.log.Info().
			Msgf("new incoming connection: %+v", conn.RemoteAddr())
		go prt.handleConn(conn)
	}
}

func (prt *TCPTransport) handleConn(conn net.Conn) {
	var err error
	defer func() {
		prt.log.Err(err).Msg("dropping peer connection.")
		conn.Close()
	}()
	peer := NewTCPPeer(conn, true)

	if err = prt.opts.HandshakeFn(peer); err != nil {
		return
	}
	if prt.opts.OnPeer != nil {
		if err = prt.opts.OnPeer(peer); err != nil {
			return
		}
	}
	var rpc RPC
	for {
		if err = prt.opts.Decoder.Decode(conn, &rpc); err != nil {
			switch {
			case errors.Is(err, net.ErrClosed):
				prt.log.Err(err).Msg("client connection closed")
				return
			case err == io.EOF:
				prt.log.Error().Msgf("client %s disconnected", conn.RemoteAddr())
				return
			}
			prt.log.Error().Msgf("TCP read error: %s", err)
		}
		rpc.From = conn.RemoteAddr()
		prt.rpcch <- rpc
	}
}
