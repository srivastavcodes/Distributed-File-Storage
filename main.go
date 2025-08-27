package main

import (
	"distorage/p2p"
	"log"
)

func main() {
	opts := p2p.TCPTransportOpts{
		ListenAddr:  ":4000",
		HandshakeFn: p2p.NOPHandshakeFn,
		Decoder:     p2p.DefaultDecoder{},
	}
	trt := p2p.NewTCPTransport(opts)
	if err := trt.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
