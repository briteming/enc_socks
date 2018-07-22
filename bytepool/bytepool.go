package bytepool

import "sync"

type BytePool struct {
    pool sync.Pool
}

func NewPool(sz int) *BytePool {
    return &BytePool{pool : sync.Pool{New:func() interface{} { return make([]byte, sz) }}}
}

func(this *BytePool) Get() []byte {
    return this.pool.Get().([]byte)
}

func(this *BytePool) Put(d []byte) {
    this.pool.Put(d)
}