package librlcom

import (
	"encoding/binary"
	"net"
)

type Connection struct {
	net.Conn
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn: conn,
	}
}

func (c *Connection) ReadBytes() ([]byte, error) {
	var err error
	var size uint64
	err = binary.Read(c.Conn, binary.LittleEndian, &size)
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, size)
	_, err = c.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func (c *Connection) WriteBytes(p []byte) error {
	var err error
	var size uint64 = uint64(len(p))
	err = binary.Write(c.Conn, binary.LittleEndian, &size)
	if err != nil {
		return err
	}
	_, err = c.Write(p)
	if err != nil {
		return err
	}
	return nil
}
