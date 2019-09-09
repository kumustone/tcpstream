package tcpstream

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	Magic     = 0xAD12
	HeaderLen = 18
)

var reqID Int

type Message struct {
	Header ProtoHeader
	Body   []byte
}

func (m *Message) SetSeqID(seq uint64) {
	m.Header.Seq = seq
}

func (m *Message)Decode(buffer []byte) error {
	if len(buffer) < HeaderLen {
		return errors.New("buffer len < HeaderLen")
	}

	if err := m.Header.decode(buffer); err != nil {
		return err
	}
	fmt.Println("msg Header len ", m.Header.len)
	if len(buffer) < int(HeaderLen + m.Header.len){
		return errors.New("buffer len is not enough")
	}

	m.Body = buffer[HeaderLen:]

	return nil
}

func (m *Message) Encode() []byte {
	return m.Header.encode(m.Body)
}

type ProtoHeader struct {
	magic   uint16
	len     uint32
	Seq     uint64
	MsgType uint32
}

func (p *ProtoHeader) decode(buf []byte) error {
	p.magic = binary.BigEndian.Uint16(buf[0:2])
	if p.magic != Magic {
		return errors.New("Magic is invalid ")
	}

	p.len = binary.BigEndian.Uint32(buf[2:6])

	if p.len > MaxMsgSize {
		return errors.New("msg len extend max size ")
	}

	p.Seq = binary.BigEndian.Uint64(buf[6:14])

	p.MsgType = binary.BigEndian.Uint32(buf[14:18])

	return nil
}

func (p *ProtoHeader) encode(input []byte) []byte {
	p.magic = Magic
	p.len = uint32(len(input))

	//解决溢出为0的情况
	//p.Seq = reqID.GenID()

	buffers := make([]byte, HeaderLen+len(input))

	binary.BigEndian.PutUint16(buffers[0:2], p.magic)
	binary.BigEndian.PutUint32(buffers[2:6], p.len)
	binary.BigEndian.PutUint64(buffers[6:14], p.Seq)
	binary.BigEndian.PutUint32(buffers[14:18], p.MsgType)

	copy(buffers[18:], input)

	return buffers
}
