package tests

import (
    "testing"
    "net"
    "encoding/binary"
    "fmt"
    "encoding/hex"
)

func CheckLen(data []byte) int {
    if len(data) <= 4 {
        return 0
    }
    fmt.Printf("Recv data len:%d, hex:%s\n", len(data), hex.EncodeToString(data))
    xlen := binary.BigEndian.Uint32(data)
    if len(data) < int(xlen) {
        return 0
    }
    if xlen > 64 * 1024 {
        return -1
    }
    return int(xlen)
}

func handleConnect(conn net.Conn) {
    defer func() {
        conn.Close()
    }()
    buf := make([]byte, 64 * 1024)
    readIndex := 0
    fmt.Printf("Accept connect:%s\n", conn.RemoteAddr().String())
    for ; ;{
        cnt, err := conn.Read(buf[readIndex:])
        if err != nil {
            fmt.Printf("Read data failed, e:%s\n", err.Error())
            return
        }
        readIndex += cnt
        xlen := CheckLen(buf[0:readIndex])
        if xlen == 0 {
            fmt.Printf("No much data found, len:%d, data:%s, datalen:%d\n", xlen, string(buf[0:readIndex]), readIndex)
            continue
        }
        if xlen < 0 {
            fmt.Printf("Read data failed, len:%d", xlen)
            return
        }
        data := make([]byte, xlen)
        copy(data, buf[:xlen])
        fmt.Printf("Read data:%s, len:%d, hex:%s from client, send back!\n", string(data[4:]), len(data[4:]), hex.EncodeToString(data))
        if xlen < readIndex {
            copy(buf, buf[xlen:readIndex])
            readIndex -= xlen
        } else {
            readIndex = 0
        }
        cnt, err = conn.Write(data)
        if cnt != len(data) || err != nil {
            fmt.Printf("Write data back failed, err:%v, cnt:%d\n", err, cnt)
            return
        }
        return
    }
}



func TestEchoSvr(t *testing.T) {
    addr := "127.0.0.1:8849"
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        t.Errorf("Echo svr listen failed, err:%s", err.Error())
        return
    }
    fmt.Printf("Echo svr listen on %s\n", addr)
    for {
        conn, err := listener.Accept()
        if err != nil {
            t.Errorf("Accept failed, err:%s", err.Error())
            return
        }
        go func() {
            handleConnect(conn)
        }()
    }
}

