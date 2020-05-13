package blaster

type cmsgPair struct {
	h cmsg
	p interface{}
}

type cmsg struct {
	seq int32
	mt  cmsgType
}

type cmsgType uint8

const (
	Hello cmsgType = iota
	Close
)

type chello struct {
	nonce string
}