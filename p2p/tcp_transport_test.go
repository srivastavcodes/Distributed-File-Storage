package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr:  ":4000",
		HandshakeFn: NOPHandshakeFn,
	}
	trt := NewTCPTransport(opts)
	assert.Equal(t, trt.ListenAddr, opts.ListenAddr)

	assert.NoError(t, trt.ListenAndAccept())
}
