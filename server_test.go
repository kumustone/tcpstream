package tcpstream

import "testing"

type serverHandler struct {
}

func (h *serverHandler) OnData(conn *TcpStream, msg *Message) error {
	_ = conn.Write(&Message{
		Header: ProtoHeader{
			Seq: msg.Header.Seq,
		},
		Body: []byte("I am server"),
	})
	return nil
}
func (h *serverHandler) OnConn(conn *TcpStream)    {}
func (h *serverHandler) OnDisConn(conn *TcpStream) {}

func TestSyncServer_Call(t *testing.T) {
	s := NewTCPServer("127.0.0.1:7001", &handler{})
	s.Serve()
}
