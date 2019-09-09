package tcpstream

import (
	"errors"
	"fmt"
	"net"
)

type TCPServer struct {
	Address string
	Handler StreamHandler
	Manager *Manager

	listener net.Listener
}

func NewTCPServer(listenAddr string, h StreamHandler) *TCPServer {
	return &TCPServer{
		Address: listenAddr,
		Handler: h,
		Manager: NewManager(),
	}
}

func (t *TCPServer) Serve() error {
	l, err := net.Listen("tcp4", t.Address)
	if err != nil {
		return err
	}

	var e error
	t.listener = l
	go func() {
		for {
			tcpConn, err := t.listener.Accept()
			if err != nil {
				e = err
				break
			}

			stream := &TcpStream{
				ID:         GenConnID(),
				ConnType:   CONN_TYPE_SERVER,
				LocalAddr:  tcpConn.LocalAddr().String(),
				RemoteAddr: tcpConn.RemoteAddr().String(),
				State:      CONN_STATE_ESTAB,
				Conn:       tcpConn,
				Handler:    t.Handler,
				stopChan:   make(chan struct{}),
			}
			go stream.Start()
		}
	}()
	return e
}

func (t *TCPServer) SendData(id uint64, msg *Message) error {
	conn := t.Manager.Get(id)
	if conn == nil {
		return errors.New(fmt.Sprint(id, " connection is not exist!"))
	}

	return conn.Write(msg)
}

func (t *TCPServer) Stop() {
	t.listener.Close()
	t.Manager.Dispose()
}
