package p2p

type HandshakeFunc func(peer Peer) error

func NOPHandshakeFn(_ Peer) error { return nil }
