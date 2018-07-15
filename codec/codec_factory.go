package codec

import (
    "sync"
    "errors"
)

type CodecFactory interface {
    Create() Codec
}

type CommonFactory struct {
    mp map[string]CodecFactory
    keys []string
}

var factory *CommonFactory
var mutex sync.Mutex

func NewFactory() *CommonFactory {
    if factory == nil {
        mutex.Lock()
        factory = &CommonFactory{mp:make(map[string]CodecFactory)}
        mutex.Unlock()
    }
    return factory
}

func(this *CommonFactory) Regist(name string, factory CodecFactory) {
    this.mp[name] = factory
    this.keys = append(this.keys, name)
}

func(this *CommonFactory) AllKey() []string {
    return this.keys
}

func(this *CommonFactory) GetCodec(name string) (Codec, error) {
    fac, ok := this.mp[name]
    if !ok || fac == nil {
        return nil, errors.New("not founc codec factory:" + name)
    }
    return fac.Create(), nil
}

func(this *CommonFactory) GetFactory(name string) (CodecFactory, error) {
    fac, ok := this.mp[name]
    if !ok || fac == nil {
        return nil, errors.New("not founc codec factory:" + name)
    }
    return fac, nil
}