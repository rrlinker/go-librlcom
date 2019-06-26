package librlcom

import (
	"encoding/binary"
	"net"
	"reflect"
)

type RawCourier struct {
	conn *Connection
}

func NewRawCourier(conn net.Conn) *RawCourier {
	return &RawCourier{
		conn: NewConnection(conn),
	}
}

func (rc *RawCourier) readHeader() (header Header, err error) {
	err = binary.Read(rc.conn, binary.LittleEndian, &header)
	return
}

func (rc *RawCourier) Close() error {
	return rc.conn.Close()
}

func (rc *RawCourier) Receive() (Message, error) {
	var header Header
	_, err := header.ReadFrom(rc.conn)
	if err != nil {
		return nil, err
	}
	if t, ok := typeMapMsg2Go[header.MessageType]; ok {
		msg := reflect.New(t).Interface().(Message)
		_, err = msg.ReadFrom(rc.conn)
		if err != nil {
			return &header, err
		}
		return msg, nil
	}
	return &header, ErrUnknownMessage
}

func (rc *RawCourier) Send(msg Message) error {
	if t, ok := typeMapGo2Msg[reflect.TypeOf(msg)]; ok {
		var header Header
		header.MessageType = t
		_, err := header.WriteTo(rc.conn)
		if err != nil {
			return err
		}
		_, err = msg.WriteTo(rc.conn)
		if err != nil {
			return err
		}
		return nil
	}
	return ErrUnknownMessage
}
