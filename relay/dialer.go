package relay

import (
    "net"
    "time"
    "errors"
)

func DialWithConn(conn net.Conn, msg *AuthMsg, timeout time.Duration) (net.Conn, error) {
    pkt := CreateRelayAuthReq(msg)
    if pkt == nil {
        return conn, errors.New("auth msg create failed")
    }
    err := pkt.Encode()
    if err != nil {
        return conn, err
    }
    data := pkt.GetPacketData()
    dataTotal := len(data)
    writeIndex := 0
    conn.SetWriteDeadline(time.Now().Add(timeout))
    for ; writeIndex < dataTotal; {

        cnt, err := conn.Write(data[writeIndex:])
        if err != nil {
            return conn, errors.New("write auth data failed, err:" + err.Error())
        }
        writeIndex += cnt
    }
    conn.SetWriteDeadline(time.Time{})
    return conn, nil
}

func DialTimeout(network, address string, timeout time.Duration, msg* AuthMsg) (net.Conn, error) {
    conn, err := net.DialTimeout(network, address, timeout)
    if err != nil {
        return nil, err
    }
    return DialWithConn(conn, msg, timeout)
}