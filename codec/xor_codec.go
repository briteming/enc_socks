package codec

func init() {
    c := &XorCodecFactory{}
    NewFactory().Regist("xor", c)
}

type XorCodec struct {
    key string
}

func(this *XorCodec) Encode(data []byte) ([]byte, error) {
    return this.Decode(data)
}

func(this *XorCodec) Decode(data []byte) ([]byte, error) {
    if data == nil {
        return data, nil
    }
    ret := make([]byte, len(data))
    for i := 0; i < len(data); i++ {
        ret[i] = data[i] ^ this.key[i % len(this.key)]
    }
    return ret, nil
}

func(this *XorCodec) Check(data []byte) (int, error) {
    return len(data), nil
}

func(this *XorCodec) Init(key string) error {
    this.key = key
    return nil
}

type XorCodecFactory struct {

}

func(this *XorCodecFactory) Create() Codec {
    return &XorCodec{}
}

