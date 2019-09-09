package tcpstream

type StreamHandler interface {
	OnData(conn *TcpStream, msg *Message) error
	OnConn(conn *TcpStream)
	OnDisConn(conn *TcpStream)
}

type StreamManager interface {
	Put(conn *TcpStream)
	Remove(conn *TcpStream)
}