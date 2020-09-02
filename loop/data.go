package loop

import (
	"crypto/sha512"
	"github.com/michaelquigley/dilithium/util"
	"github.com/pkg/errors"
	"math/rand"
	"time"
)

type dataBlock struct {
	data   []byte
	hash   []byte
	buffer *buffer
}

func newDataBlock(b *buffer) (*dataBlock, error) {
	for i := 4 + 64; i < int(b.sz); i++ {
		b.data[i] = byte(rand.Intn(255))
	}
	hash := sha512.Sum512(b.data[4+64:])
	copy(b.data[4:], hash[:])
	util.WriteInt32(b.data, 64)
	return &dataBlock{
		data:   b.data[4+64:],
		hash:   b.data[4:],
		buffer: b,
	}, nil
}

func (self *dataBlock) encodeDataBlock(b *buffer) error {
	required := 4 + len(self.hash) + len(self.data)
	if b.sz < uint32(required) {
		return errors.Errorf("buffer too small [%d required]", required)
	}
	util.WriteInt32(b.data[0:4], int32(len(self.hash)))
	copy(b.data[4:], self.hash)
	copy(b.data[4+len(self.hash):], self.data)
	b.uz = uint32(required)
	return nil
}

func decodeDataBlock(b *buffer) (*dataBlock, error) {
	if b.uz < 4 {
		return nil, errors.Errorf("buffer too small [at least 4 required]")
	}
	headerLen := util.ReadInt32(b.data[0:4])
	if b.uz < uint32(4+headerLen) {
		return nil, errors.Errorf("buffer too small [at least %d required]", 4+headerLen)
	}
	return &dataBlock{
		data:   b.data[4+headerLen:],
		hash:   b.data[4 : 4+headerLen],
		buffer: b,
	}, nil
}

type dataSet struct {
	blocks []*dataBlock
	pool   *pool
}

func newDataSet(sz int) (*dataSet, error) {
	ds := &dataSet{
		pool: newPool(uint32(4 + 64 + sz)),
	}

	rand.Seed(time.Now().UnixNano())
	block, err := newDataBlock(ds.pool.get())
	if err != nil {
		return nil, err
	}
	ds.blocks = append(ds.blocks, block)

	return ds, nil
}