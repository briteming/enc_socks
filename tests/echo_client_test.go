package tests

import (
    "testing"
    "net"
    "time"
    "fmt"
    "encoding/binary"
    "encoding/hex"
)

func Once(t *testing.T, i int) {
    word := "hello test, are you ok1111111111111111?-->" + fmt.Sprintf("%d\n", i)
    addr := "127.0.0.1:8847"
    conn, err := net.DialTimeout("tcp", addr, 2 * time.Second)
    if err != nil {
        t.Errorf("Dial err, addr:%s, err:%s", addr, err.Error())
        return
    }
    defer func() {
        conn.Close()
    }()
    fmt.Printf("Dial to :%s success\n", addr)
    totalLen := 4 + len(word)
    data := make([]byte, 4 + len(word))
    binary.BigEndian.PutUint32(data, uint32(totalLen))
    copy(data[4:], word)
    {
        totalLen := len(data)
        writeIndex := 0
        for ; writeIndex < totalLen; {
            writeLen, err := conn.Write(data[writeIndex:])
            if err != nil {
                fmt.Errorf("Write data to svr failed, err:%s, writeLen:%d\n", err, writeLen)
                return
            }
            writeIndex += writeLen
        }
    }
    fmt.Printf("Write data success, begin read, id:%d, hex:%s\n", i, hex.EncodeToString(data))
    result := make([]byte, len(data))
    conn.Read(result)
    fmt.Printf("Send data :%s, hex:%s, recv data:%s\n", word, hex.EncodeToString(data), string(result[4:]))
}

func TestEchoClient(t *testing.T) {
    for i := 0; i < 10000; i++ {
        Once(t, i)
        //if (i + 1) % 100 == 0 {
        //    time.Sleep(2 * time.Second)
        //}
    }
}
