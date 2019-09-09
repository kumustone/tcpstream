package tcpstream

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	ReadBufferSize = 1 << 12
	MaxMsgSize     = 1 << 23
)

const (
	CONN_STATE_LISTEN = iota
	CONN_STATE_ESTAB
	CONN_STATE_CLOSE
	CONN_STATE_CONNETING
	CONN_STATE_STOP  //主动停止掉的，不应该占用任何资源
)

const (
	CONN_TYPE_CONN   = iota //主动发起连接的TCP stream
	CONN_TYPE_SERVER        //Server派生出来的 Tcp stream
)

var connID Int

func GenConnID() uint64 {
	return connID.GenID()
}

type TcpStream struct {
	ID         uint64
	ConnType   uint8
	LocalAddr  string
	RemoteAddr string
	State      uint8
	Conn       net.Conn
	Handler    StreamHandler
	Manager    StreamManager
	stopChan   chan struct{}
	once       sync.Once
}

func NewConnTcpStream(address string, h StreamHandler) *TcpStream {
	return &TcpStream{
		ID:         GenConnID(),
		ConnType:   uint8(CONN_TYPE_CONN),
		RemoteAddr: address,
		State:      CONN_STATE_CLOSE,
		Handler:    h,
		stopChan:   make(chan struct{}),
	}
}

func (t *TcpStream) Start() {
	t.once.Do(
		func() {
			if t.ConnType == CONN_TYPE_CONN {
				t.connect()
			} else if t.ConnType == CONN_TYPE_SERVER {
				t.readLoop()
			}
		})
}

func (t *TcpStream) read() (err error) {
	head := make([]byte, HeaderLen)
	for {
		var msg Message

		//if err := t.Conn.SetReadDeadline(time.Now().Add(2 * time.Minute)); err != nil {
		//	return
		//}

		if _, err := io.ReadFull(t.Conn, head); err != nil {
			return err
		}
		if err = msg.Header.decode(head); err != nil {
			return
		}

		msg.Body = make([]byte, msg.Header.len)
		if _, err = io.ReadFull(t.Conn, msg.Body); err != nil {
			return
		}
		if t.Handler != nil {
			t.Handler.OnData(t, &msg)
		}
	}
}

func (t *TcpStream) readLoop() {
	select {
	case <-t.stopChan:
		return
	default:
		if err := t.read(); err != nil {
			fmt.Println("read err", err.Error())
			if t.Handler != nil {
				t.Handler.OnDisConn(t)
			}
			t.Conn.Close()
			t.State = CONN_STATE_CLOSE
		}
	}

	if t.ConnType == CONN_TYPE_CONN && t.State != CONN_STATE_STOP {
		go t.connect()
	}

	fmt.Println("###### (t *TcpStream) readLoop - 3")
}

//写错误不触发关闭，只有读事件触发；
func (t *TcpStream) Write(msg *Message) error {
	if t.State == CONN_STATE_ESTAB {
		_, err := t.Conn.Write(msg.Encode())
		return err
	}
	return nil
}

func (t *TcpStream) connect() error {
	if t.State == CONN_STATE_CONNETING || t.State == CONN_STATE_ESTAB {
		return errors.New("already connected")
	}

	t.State = CONN_STATE_CONNETING
	for {
		c, err := net.DialTimeout("tcp", t.RemoteAddr, 5*time.Second)
		if err != nil {
			fmt.Println("Connect fail ", err.Error())

			timer := time.NewTimer(time.Second * 5)
			select {
			case <-timer.C:
				continue
			case <-t.stopChan:
				return nil
			}
		} else {
			fmt.Println("Connect success !!!!")
			t.Conn = c
			t.State = CONN_STATE_ESTAB
			if t.Handler != nil {
				t.Handler.OnConn(t)
			}
			if t.Manager != nil {
				t.Manager.Put(t)
			}

			go t.readLoop()
			break
		}
	}
	return nil
}

func (t *TcpStream) Stop() {
	close(t.stopChan)
	t.Conn.Close()
	t.State = CONN_STATE_STOP
}
