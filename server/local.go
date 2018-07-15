package server

import (
    "time"
    "net"
    log "github.com/sirupsen/logrus"
    "github.com/xxxsen/enc_socks/codec"
    "sync"
    "io"
    "github.com/xxxsen/enc_socks/relay_msg"
    "context"
    "github.com/xxxsen/enc_socks/packet"
)

type RelayLocal struct {
    config *ServerConfig
}

func NewRelayLocal(cfg *ServerConfig) *RelayLocal {
    return &RelayLocal{config:cfg}
}

func(this *RelayLocal) Start() {
    listener, err := net.Listen("tcp", this.config.LocalAddr)
    if err != nil {
        log.Errorf("Listen addr:%s failed, err:%s", this.config.LocalAddr, err.Error())
        return
    }
    var sessionId uint32 = 0;
    log.Printf("Local server start on addr:%s", this.config.LocalAddr)
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Errorf("Rcv conn failed, err:%s", err)
            time.Sleep(100 * time.Millisecond)
            continue
        }
        sessionId++
        log.Infof("Recv connection from local, addr:%s, mark sessionid:%d", conn.RemoteAddr(), sessionId)
        go func() {
            this.handleTcpConnection(conn, sessionId)
        }()
    }
}

func(this *RelayLocal) handleTcpConnection(local net.Conn, sessionId uint32) {
    remote, err := net.DialTimeout("tcp", this.config.RemoteAddr, this.config.Timeout)
    if err != nil {
        log.Errorf("Dial remote svr failed, err:%s, addr:%s,sessionid:%d", err.Error(), this.config.RemoteAddr, sessionId)
        return
    }
    defer func() {
        local.Close()
        remote.Close()
    }()
    fac, cerr := codec.NewFactory().GetFactory(this.config.Codec)
    if cerr != nil {
        log.Errorf("Get codec factory failec, name:%s, sessionid:%d", this.config.Codec, sessionId)
        return
    }
    ctx, cancel := context.WithCancel(context.Background())
    waiter := sync.WaitGroup{}
    waiter.Add(2)
    log.Infof("Connection establish, local:%s, remote:%s, sessionid:%d, transfer data!",
        local.RemoteAddr().String(), remote.RemoteAddr().String(), sessionId)
    go func() {  //read remote write local
        defer func() {
            waiter.Done()
            cancel()
        }()
        buf := make([]byte, 128 * 1024)
        readIndex := 0
        coder := fac.Create()
        coder.Init(this.config.Key)
        pkt := packet.NewRelayPacket(coder)
        for {
            remote.SetReadDeadline(time.Now().Add(this.config.Timeout))
            cnt, err := remote.Read(buf[readIndex:])
            if err != nil {
                if err == io.EOF {
                    log.Printf("Remote read eof, exit, addr:%s, session:%d", remote.RemoteAddr().String(), sessionId)
                    return ;
                } else if err, ok := err.(net.Error); ok && err.Timeout() {
                    if(isDone(ctx)) {
                        log.Errorf("Local thread already exit, quit remote read, remte:%s, session:%d",
                            remote.RemoteAddr().String(), sessionId)
                        return
                    }
                    continue
                } else {
                    log.Errorf("Remote read err, err:%s, addr:%s, session:%d", err, remote.RemoteAddr().String(), sessionId)
                    return
                }
            }
            readIndex += cnt
            for ;; {
                checkLen, checkErr := packet.CheckRelayCodec(buf[0:readIndex], coder)
                if checkErr != nil {
                    log.Errorf("Check remote packet failed, checkResult:%d, err:%s, sessionid:%d", checkLen, checkErr.Error(), sessionId)
                    return
                }
                if checkLen == 0 {
                    if readIndex != 0 {
                        log.Debugf("Check remote packet, but need more data, spare data len:%d, sessionid:%d", readIndex, sessionId)
                    }
                    break
                }
                pkt.SetData(buf[0:checkLen])
                dret, derr := pkt.Decode()
                if derr != nil {
                    log.Errorf("Decode remote packet failed, ret:%d, err:%s, data len:%d, sessionid:%d",
                        dret, derr.Error(), checkLen, sessionId)
                    return
                }
                if(readIndex > checkLen) {
                    copy(buf, buf[checkLen:readIndex])
                    readIndex = readIndex - checkLen
                } else {
                    readIndex = 0
                }
                if pkt.GetMsgCmd() == int32(relay_msg.CMD_TYPE_CMD_PAD) {
                    continue
                }
                {
                    writeData := pkt.GetMsgData()
                    writeIndex := 0
                    totalData := len(writeData)
                    for ; writeIndex < totalData; {
                        writeLen, werr := local.Write(writeData[writeIndex:])
                        if werr != nil {
                            log.Errorf("Write remote data to local failed, ret:%d, err:%s, sessionid:%d", writeLen, werr.Error(), sessionId)
                            return ;
                        }
                        writeIndex += writeLen
                    }
                }
            }
            if(isDone(ctx)) {
                log.Printf("Remote read exit cause by other cancel, sessionid:%d", sessionId)
                return ;
            }
        }
    }()

    go func() {  //read local write remote
        defer func() {
            waiter.Done()
            cancel()
        }()
        buf := make([]byte, 64 * 1024)
        readIndex := 0
        coder := fac.Create()
        coder.Init(this.config.Key)
        pkt := packet.NewRelayPacket(coder)

        for {
            local.SetReadDeadline(time.Now().Add(this.config.Timeout))
            cnt, err := local.Read(buf[readIndex:])
            if err != nil {
                if err == io.EOF {
                    log.Printf("Local read eof, exit, addr:%s, session:%d", local.RemoteAddr().String(), sessionId)
                    return ;
                } else if err, ok := err.(net.Error); ok && err.Timeout() {
                    if(isDone(ctx)) {
                        log.Errorf("Remote thread already exit, quit local read! session:%d", sessionId)
                        return
                    }
                    continue
                } else {
                    log.Errorf("Local read err, err:%s, addr:%s, session:%d", err, local.RemoteAddr().String(), sessionId)
                    return
                }
            }
            readIndex += cnt
            pkt.SetMsgCmd(int32(relay_msg.CMD_TYPE_CMD_DATA))
            pkt.SetMsgData(buf[0:readIndex])
            eret, eerr := pkt.Encode()
            if eerr != nil {
                log.Errorf("Encode local pkt failed, ret:%d, err:%s, session:%d", eret, eerr.Error(), sessionId)
                return
            }
            readIndex = 0
            {
                writeIndex := 0
                writeData := pkt.GetData()
                writeTotal := len(writeData)
                for ; writeIndex < writeTotal; {
                    writeLen, werr := remote.Write(writeData[writeIndex:])
                    if werr != nil {
                        log.Errorf("Write local data to remote failed, ret:%d, werr:%s, session:%d", writeLen, werr.Error(), sessionId)
                        return
                    }
                    writeIndex += writeLen
                }
            }
            if(isDone(ctx)) {
                log.Errorf("Local read exit cause by other cancel, session:%d", sessionId)
            }
        }
    }()
    waiter.Wait()
    log.Printf("Process connect finish, local:%s, remote:%s, session:%d", local.RemoteAddr().String(), remote.RemoteAddr().String(), sessionId)
}