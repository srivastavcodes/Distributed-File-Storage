package p2p

type HandshakeFunc func(peer Peer) error

func NOPHandshakeFn(peer Peer) error { return nil }
