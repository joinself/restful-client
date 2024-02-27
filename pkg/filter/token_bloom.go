package filter

import (
	"encoding/binary"

	"github.com/bits-and-blooms/bloom/v3"
)

type Checker struct {
	f *bloom.BloomFilter
}

func NewChecker() *Checker {
	return &Checker{
		f: bloom.New(1000, 4),
	}
}

func (c *Checker) Add(val int) {
	n := make([]byte, 4)
	binary.BigEndian.PutUint32(n, uint32(val))
	c.f.Add(n)
}

func (c *Checker) Check(val int) bool {
	n := make([]byte, 4)
	binary.BigEndian.PutUint32(n, uint32(val))
	return c.f.Test(n)
}
