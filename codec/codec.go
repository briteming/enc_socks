package codec

type Codec interface {
    Encode([]byte) ([]byte, error)
    Decode([]byte) ([]byte, error)
    Check([]byte) (int, error)
    Init(key string) error
}

