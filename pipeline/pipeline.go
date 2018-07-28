package pipeline

import (
    "net"
    "errors"
    "fmt"
    "time"
)

type Pipeline interface {
    Process(net.Conn) (net.Conn, error)
}

type Factory interface {
    GetName() string
    GetSvrPipe() Pipeline
    GetCliPipe() Pipeline
    Init(mp *map[string]interface{}) error
}

type FactoryHolder struct {
    mp map[string]Factory
}

type DefaultConn struct {
    Conn net.Conn
}

func(this *DefaultConn) Read(b []byte) (int, error) {
    return this.Conn.Read(b)
}

func(this *DefaultConn) Write(b []byte) (int, error) {
    return this.Conn.Write(b)
}

func(this *DefaultConn) Close() error {
    return this.Conn.Close()
}

func(this *DefaultConn) LocalAddr() net.Addr {
    return this.Conn.LocalAddr()
}

func(this *DefaultConn) RemoteAddr() net.Addr {
    return this.Conn.RemoteAddr()
}

func(this *DefaultConn) SetDeadline(t time.Time) error {
    return this.SetDeadline(t)
}

func(this *DefaultConn) SetReadDeadline(t time.Time) error {
    return this.Conn.SetReadDeadline(t)
}


func(this *DefaultConn) SetWriteDeadline(t time.Time) error {
    return this.Conn.SetWriteDeadline(t)
}


var holder FactoryHolder = FactoryHolder{ mp : make(map[string]Factory) }

func GetHolder() *FactoryHolder {
    return &holder
}

func(this *FactoryHolder) Regist(factory Factory) error {
    if _, ok := this.mp[factory.GetName()]; ok {
        return errors.New(fmt.Sprintf("factory:%s already regist", factory.GetName()))
    }
    this.mp[factory.GetName()] = factory
    return nil
}

func(this *FactoryHolder) GetByName(name string) (Factory, error) {
    if fac, ok := this.mp[name]; ok {
        return fac, nil
    } else {
        return nil, errors.New(fmt.Sprintf("factory:%s not found", name))
    }
}