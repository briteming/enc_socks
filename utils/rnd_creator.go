package utils

import (
    "sync"
    "math/rand"
    "time"
)

type RandCreator struct {
    rnd *rand.Rand
    table string
}

var rndc *RandCreator
var mutex sync.Mutex

func NewRand() *RandCreator {
    if rndc == nil {
        mutex.Lock()
        if rndc == nil {
            rndc = &RandCreator{ rnd:rand.New(rand.NewSource(time.Now().UnixNano())), table: "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" }
            rndc.rnd.Seed(time.Now().Unix())
        }
        mutex.Unlock()
    }
    return rndc
}

func(this *RandCreator) CreateBytes(sz int) []byte {
    mutex.Lock()
    data := make([]byte, sz)
    for i := 0; i < sz; i++ {
        data[i] = this.table[this.rnd.Int() % len(this.table)]
    }
    mutex.Unlock()
    return data
}

func(this *RandCreator) CreateUInt32(sz uint32) uint32 {
    if sz == 0 {
        return 0
    }
    mutex.Lock()
    v := this.rnd.Uint32() % sz
    mutex.Unlock()
    return v;
}