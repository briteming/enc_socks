package pipeline

import "net"

type XorPipeline struct {
    mp *map[string]interface{}
}

type XorConn struct {
    *DefaultConn
}

func(this *XorPipeline) Process(conn net.Conn) (net.Conn, error) {
    return &XorConn{&DefaultConn{conn}}, nil
}

func(this *XorConn) xor(b []byte) {
    for i := 0; i < len(b); i++ {
        b[i] = b[i] ^ 0xff
    }
}

func(this *XorConn) Read(b []byte) (int, error) {
    cnt, err := this.Conn.Read(b);
    if err != nil {
        return cnt, err
    }
    this.xor(b[:cnt])
    return cnt, err
}

func(this *XorConn) Write(b []byte) (int, error) {
    tmp := make([]byte, len(b))
    copy(tmp, b)
    this.xor(tmp)
    return this.Conn.Write(tmp)
}

type XorFactory struct {
    mp *map[string]interface{}
}

func(this *XorFactory) GetName() string {
    return "xor"
}

func(this *XorFactory) GetSvrPipe() Pipeline {
    return &XorPipeline{ this.mp }
}

func(this *XorFactory) GetCliPipe() Pipeline {
    return &XorPipeline{this.mp }
}

func(this *XorFactory) Init(mp *map[string]interface{}) error {
    this.mp = mp
    return nil
}

func NewXorFactory() *XorFactory {
    return &XorFactory{}
}

func init() {
    GetHolder().Regist(NewXorFactory())
}