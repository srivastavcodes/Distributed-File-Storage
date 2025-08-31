package main

import (
	"distorage/p2p"
	"distorage/server"
	store "distorage/storage"
	"time"
)

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:  ":3000",
		HandshakeFn: p2p.NOPHandshakeFn,
		Decoder:     p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := server.FileServerOpts{
		ListenAddr:      tcpTransportOpts.ListenAddr,
		StorageRoot:     "3000_network",
		Transport:       tcpTransport,
		PathTransformFn: store.CASPathTransformFunc,
	}
	srv := server.NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(time.Second * 5)
		srv.Stop()
	}()

	if err := srv.Start(); err != nil {
		srv.Log.Fatal().Msgf("server stopped abruptly: %s", err)
	}
}
