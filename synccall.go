package tcpstream

import (
	"errors"
	"time"
)

/*
Conn : 使用rpc调用通道;
msg : 发送的数据
timeout: 调用超时时间
resp : resp数据
*/

type SyncClient struct {
	reqCache *RequestManager
	stream   *TcpStream
}

func NewSyncClient(address string) *SyncClient {
	c := &SyncClient{
		reqCache: NewRequestManager(),
	}
	c.stream = NewConnTcpStream(address, c)
	c.stream.Start()
	return c
}

func NewSyncFromTcpStream(s *TcpStream)  *SyncClient{
	return &SyncClient{
		stream : s,
		reqCache: NewRequestManager(),
	}
}

func (c *SyncClient) OnData(conn *TcpStream, msg *Message) error {
	return c.reqCache.shoot(msg.Header.Seq, msg)
}

func (c *SyncClient) OnConn(conn *TcpStream) {
	return
}

func (c *SyncClient) OnDisConn(conn *TcpStream) {
	return
}

func (c *SyncClient) Call(msg *Message, timeOut time.Duration) (resp *Message, err error) {
	if c.stream.State != CONN_STATE_ESTAB {
		err = errors.New("connection is not exist")
		return
	}
	reqID := reqID.GenID()
	msg.SetSeqID(reqID)

	respChan := make(chan *Message, 2)
	c.reqCache.cache(reqID, respChan)

	if err = c.stream.Write(msg); err != nil {
		return
	}

	timer := time.NewTimer(timeOut)
	select {
	case resp = <-respChan:
		return
	case <-timer.C:
		c.reqCache.dispose(reqID)
	}

	return nil, errors.New("SyncCall timeout")
}
