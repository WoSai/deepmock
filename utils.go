package deepmock

import (
	"crypto/md5"
	"fmt"
	"hash"
	"sync"
)

type (
	hashPool struct {
		salt string
		pool sync.Pool
	}
)

var (
	defaultHashPoll *hashPool
	salt            = []byte(`6ee30676-6c88-4d3a-86b1-bb61e82da1c9`)
)

func newHashPool() *hashPool {
	return &hashPool{
		pool: sync.Pool{New: func() interface{} {
			return md5.New()
		}}}
}

func (fp *hashPool) get() hash.Hash {
	return fp.pool.Get().(hash.Hash)
}

func (fp *hashPool) put(h hash.Hash) {
	h.Reset()
	fp.pool.Put(h)
}

func genID(path, method []byte) string {
	h := defaultHashPoll.get()
	defer defaultHashPoll.put(h)

	h.Write(method)
	h.Write(path)
	h.Write(salt)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func init() {
	defaultHashPoll = newHashPool()
}
