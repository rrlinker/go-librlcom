package librlcom

import (
	"encoding/binary"
	"errors"
	"net"
	"reflect"
)

var (
	ErrUnknownMessage = errors.New("unknown message")
)

type Courier struct {
	conn net.Conn
}

func NewCourier(conn net.Conn) *Courier {
	return &Courier{
		conn: conn,
	}
}

func (c *Courier) readHeader() (header Header, err error) {
	err = binary.Read(c.conn, binary.LittleEndian, &header)
	return
}

func (c *Courier) Receive() (Message, error) {
	var header Header
	_, err := header.ReadFrom(c.conn)
	if err != nil {
		return nil, err
	}
	if t, ok := typeMapMsg2Go[header.MessageType]; ok {
		msg := reflect.New(t).Interface().(Message)
		_, err = msg.ReadFrom(c.conn)
		if err != nil {
			return &header, err
		}
		return msg, nil
	}
	return &header, ErrUnknownMessage
}

func (c *Courier) Send(msg Message) error {
	if t, ok := typeMapGo2Msg[reflect.TypeOf(msg)]; ok {
		var header Header
		header.MessageSize = uint64(msg.Size())
		header.MessageType = t
		_, err := header.WriteTo(c.conn)
		if err != nil {
			return err
		}
		_, err = msg.WriteTo(c.conn)
		if err != nil {
			return err
		}
		return nil
	}
	return ErrUnknownMessage
}
