package misc

import (
	"bytes"
	"encoding/hex"
	"hash"
	"math/rand"
	"sync"
	"time"

	"github.com/spaolacci/murmur3"
)

type (
	hashPool struct {
		salt string
		pool sync.Pool
	}
)

const (
	letterBytes   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits
)

var (
	defaultHashPoll *hashPool
	salt            = []byte(`6ee30676-6c88-4d3a-86b1-bb61e82da1c9`)
)

func newHashPool() *hashPool {
	return &hashPool{
		pool: sync.Pool{New: func() interface{} {
			return murmur3.New32()
		}}}
}

func (fp *hashPool) get() hash.Hash {
	return fp.pool.Get().(hash.Hash)
}

func (fp *hashPool) put(h hash.Hash) {
	h.Reset()
	fp.pool.Put(h)
}

// GenID 基于murmur3的哈希函数
func GenID(path, method []byte) string {
	h := defaultHashPoll.get()
	defer defaultHashPoll.put(h)

	h.Write(bytes.ToUpper(method))
	h.Write(path)
	h.Write(salt)
	return hex.EncodeToString(h.Sum(nil))
}

// GenRandomString 生产指定长度的随机字符串
func GenRandomString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

func init() {
	// create default hash pool
	defaultHashPoll = newHashPool()
	rand.Seed(time.Now().UnixNano())
}
