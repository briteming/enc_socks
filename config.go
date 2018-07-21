package enc_socks

import (
    "time"
    "bytes"
    "fmt"
    "enc_socks/relay"
)

const (
    SERVER_TYPE_LOCAL = 1;
    SERVER_TYPE_REMOTE = 2;
)

type ServerConfig struct {
    LocalAddr  string
    RemoteAddr string
    Timeout    time.Duration
    ServerType int32
    User       relay.AuthMsg
    UserInfo   relay.AuthMap
    TlsServerPemAddr string
    TlsServerKeyAddr string
}

func(this *ServerConfig) String() string {
    var buffer bytes.Buffer
    buffer.WriteString("ServerConfig:[")
    buffer.WriteString("local:" + this.LocalAddr + ", ")
    buffer.WriteString("remote:" + this.RemoteAddr + ", ")
    buffer.WriteString("timeout:" + fmt.Sprintf("%ds, ", this.Timeout / time.Second))
    buffer.WriteString("type:"  + fmt.Sprintf("%d, ", this.ServerType))
    buffer.WriteString("tls:[Pem:" + this.TlsServerPemAddr + ", Key:" + this.TlsServerKeyAddr + "], ")
    buffer.WriteString("]")
    return buffer.String()
}
