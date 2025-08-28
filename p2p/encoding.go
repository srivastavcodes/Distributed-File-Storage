package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(reader io.Reader, rpc *RPC) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(reader io.Reader, rpc *RPC) error {
	return gob.NewDecoder(reader).Decode(rpc)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(reader io.Reader, rpc *RPC) error {
	buf := make([]byte, 16384)

	n, err := reader.Read(buf)
	if err != nil {
		return err
	}
	rpc.Payload = buf[:n]
	return nil
}
