package server

import (
    "time"
    "bytes"
    "fmt"
)

type ServerConfig struct {
    LocalAddr  string
    RemoteAddr string
    Key        string
    Timeout    time.Duration
    Codec      string
}

func(this *ServerConfig) String() string {
    var buffer bytes.Buffer
    buffer.WriteString("ServerConfig:[")
    buffer.WriteString("local:" + this.LocalAddr + ", ")
    buffer.WriteString("remote:" + this.RemoteAddr + ", ")
    buffer.WriteString("Key:" + this.Key + ", ")
    buffer.WriteString("Timeout:" + fmt.Sprintf("%ds, ", this.Timeout / time.Second))
    buffer.WriteString("Codec:" + this.Codec + ", ")
    buffer.WriteString("]")
    return buffer.String()
}
