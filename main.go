package main

import (
	"distorage/p2p"
	"log"

	"github.com/rs/zerolog"
)

func main() {
	writer := zerolog.NewConsoleWriter()
	logger := zerolog.New(writer).With().Timestamp().Logger()

	peerFn := func(peer p2p.Peer) error {
		peer.Close()
		return nil
	}
	opts := p2p.TCPTransportOpts{
		ListenAddr:  ":4000",
		HandshakeFn: p2p.NOPHandshakeFn,
		OnPeer:      peerFn,
		Decoder:     p2p.DefaultDecoder{},
	}
	trt := p2p.NewTCPTransport(opts)
	go func() {
		for {
			msg := <-trt.Consume()
			logger.Info().Msgf("%+v\n", msg)
		}
	}()
	if err := trt.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
