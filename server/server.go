package server

import (
	"distorage/p2p"
	store "distorage/storage"

	"github.com/rs/zerolog"
)

type FileServerOpts struct {
	Log             zerolog.Logger
	ListenAddr      string
	StorageRoot     string
	Transport       p2p.Transport
	PathTransformFn store.PathTransformFn
}

type FileServer struct {
	FileServerOpts

	store  *store.Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	writer := zerolog.NewConsoleWriter()
	logger := zerolog.New(writer).With().Timestamp().Logger()

	opts.Log = logger

	storeOpts := store.StoreOpts{
		Root:            opts.StorageRoot,
		PathTransformFn: opts.PathTransformFn,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          store.NewStore(storeOpts),
		quitch:         make(chan struct{}),
	}
}

func (srv *FileServer) Stop() { close(srv.quitch) }

func (srv *FileServer) loop() {
	defer func() {
		srv.Log.Info().Msgf("stopping server port=%s", srv.ListenAddr)
		srv.Transport.Close()
	}()
outer:
	for {
		select {
		case msg := <-srv.Transport.Consume():
			srv.Log.Info().Msgf("%v", msg)
		case <-srv.quitch:
			break outer
		}
	}
}

func (srv *FileServer) Start() error {
	if err := srv.Transport.ListenAndAccept(); err != nil {
		return err
	}
	// blocks, so user can start the server in a go-routine
	srv.loop()
	return nil
}
