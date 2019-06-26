package librlcom

import (
	"crypto/rsa"
	"encoding/binary"
	"net"
	"reflect"
)

type CryptoCourier struct {
	conn *CryptoConnection
}

func NewCryptoCourier(conn net.Conn) *CryptoCourier {
	return &CryptoCourier{
		conn: NewCryptoConnection(conn),
	}
}

func (cc *CryptoCourier) readHeader() (header Header, err error) {
	err = binary.Read(cc.conn, binary.LittleEndian, &header)
	return
}

func (cc *CryptoCourier) Close() error {
	return cc.conn.Close()
}

func (cc *CryptoCourier) Key() []byte {
	return cc.conn.key
}

func (cc *CryptoCourier) Receive() (Message, error) {
	var err error
	err = cc.conn.GatherAndDecrypt()
	if err != nil {
		return nil, err
	}
	var header Header
	_, err = header.ReadFrom(cc.conn)
	if err != nil {
		return nil, err
	}
	if t, ok := typeMapMsg2Go[header.MessageType]; ok {
		msg := reflect.New(t).Interface().(Message)
		_, err = msg.ReadFrom(cc.conn)
		if err != nil {
			return &header, err
		}
		return msg, nil
	}
	return &header, ErrUnknownMessage
}

func (cc *CryptoCourier) Send(msg Message) error {
	if t, ok := typeMapGo2Msg[reflect.TypeOf(msg)]; ok {
		var header Header
		header.MessageType = t
		_, err := header.WriteTo(cc.conn)
		if err != nil {
			return err
		}
		_, err = msg.WriteTo(cc.conn)
		if err != nil {
			return err
		}
		cc.conn.EncryptAndFlush()
		return nil
	}
	return ErrUnknownMessage
}

func (cc *CryptoCourier) InitAsClient(pub *rsa.PublicKey) error {
	return cc.conn.InitAsClient(pub)
}

func (cc *CryptoCourier) InitAsServer(priv *rsa.PrivateKey) error {
	return cc.conn.InitAsServer(priv)
}

func (cc *CryptoCourier) InitWithKey(key []byte) error {
	return cc.conn.InitWithKey(key)
}
