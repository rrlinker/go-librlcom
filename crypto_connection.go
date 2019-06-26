package librlcom

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"net"
)

var (
	ErrTriedToReadBeyondBuffer = errors.New("tried to read beyond read buffer")
	ErrReadBufferNotEmpty      = errors.New("read buffer is not empty")
)

type CryptoConnection struct {
	conn        *Connection
	key         []byte
	cipher      cipher.Block
	readOffset  int
	readBuffer  []byte
	writeBuffer []byte
}

func NewCryptoConnection(conn net.Conn) *CryptoConnection {
	cc := &CryptoConnection{
		conn: NewConnection(conn),
	}
	cc.resetReadBuffer()
	cc.resetWriteBuffer()
	return cc
}

func (cc *CryptoConnection) Close() error {
	return cc.conn.Close()
}

func (cc *CryptoConnection) Key() []byte {
	return cc.key
}

func (cc *CryptoConnection) Read(p []byte) (n int, err error) {
	n = len(p)
	if n > len(cc.readBuffer)-cc.readOffset {
		n = len(cc.readBuffer) - cc.readOffset
		if n < 0 {
			n = 0
		}
		err = ErrTriedToReadBeyondBuffer
	}
	if n > 0 {
		copy(p, cc.readBuffer[cc.readOffset:cc.readOffset+n])
		cc.readOffset += n
	}
	if cc.readOffset >= len(cc.readBuffer) {
		cc.readBuffer = nil
	}
	return
}

func (cc *CryptoConnection) Write(p []byte) (n int, err error) {
	cc.writeBuffer = append(cc.writeBuffer, p...)
	return len(p), nil
}

func (cc *CryptoConnection) InitAsClient(pub *rsa.PublicKey) error {
	var err error
	cc.resetWriteBuffer()
	key := make([]byte, 16)
	_, err = rand.Read(key)
	_, err = cc.Write(key)
	if err != nil {
		return err
	}
	cc.setWriteBufferSize()
	cc.writeBuffer, err = rsa.EncryptOAEP(sha1.New(), rand.Reader, pub, cc.writeBuffer, nil)
	if err != nil {
		return err
	}
	err = cc.conn.WriteBytes(cc.writeBuffer)
	if err != nil {
		return err
	}
	cc.resetWriteBuffer()
	return cc.InitWithKey(key)
}

func (cc *CryptoConnection) InitAsServer(priv *rsa.PrivateKey) error {
	var err error
	err = cc.EnsureReadBufferEmpty()
	if err != nil {
		return err
	}
	cc.readBuffer, err = cc.conn.ReadBytes()
	if err != nil {
		return err
	}
	cc.readBuffer, err = rsa.DecryptOAEP(sha1.New(), nil, priv, cc.readBuffer, nil)
	if err != nil {
		return err
	}
	cc.setReadBufferSize()
	key := make([]byte, len(cc.readBuffer)-cc.readOffset)
	_, err = cc.Read(key)
	if err != nil {
		return err
	}
	return cc.InitWithKey(key)
}

func (cc *CryptoConnection) InitWithKey(key []byte) error {
	var err error
	cc.key = make([]byte, len(key), len(key))
	copy(cc.key, key)
	cc.cipher, err = aes.NewCipher(key)
	if err != nil {
		return err
	}
	return nil
}

func (cc *CryptoConnection) GatherAndDecrypt() error {
	err := cc.gather()
	if err != nil {
		return err
	}
	cc.decrypt()
	return nil
}

func (cc *CryptoConnection) EncryptAndFlush() error {
	cc.encrypt()
	err := cc.flush()
	if err != nil {
		return err
	}
	return nil
}

func (cc *CryptoConnection) EnsureReadBufferEmpty() error {
	if cc.readBuffer != nil {
		return ErrReadBufferNotEmpty
	}
	return nil
}

func (cc *CryptoConnection) setWriteBufferSize() error {
	sizeBuffer := new(bytes.Buffer)
	size := uint64(len(cc.writeBuffer))
	err := binary.Write(sizeBuffer, binary.LittleEndian, size)
	if err != nil {
		return err
	}
	copy(cc.writeBuffer[0:8], sizeBuffer.Bytes())
	return nil
}

func (cc *CryptoConnection) setReadBufferSize() {
	size := uint64(cc.readBuffer[0])<<0 |
		uint64(cc.readBuffer[1])<<8 |
		uint64(cc.readBuffer[2])<<16 |
		uint64(cc.readBuffer[3])<<24 |
		uint64(cc.readBuffer[4])<<32 |
		uint64(cc.readBuffer[5])<<40 |
		uint64(cc.readBuffer[6])<<48 |
		uint64(cc.readBuffer[7])<<56
	cc.readBuffer = cc.readBuffer[:size]
	cc.readOffset = 8 // sizeof uint64
}

func (cc *CryptoConnection) gather() error {
	var err error
	err = cc.EnsureReadBufferEmpty()
	if err != nil {
		return err
	}
	cc.readBuffer, err = cc.conn.ReadBytes()
	if err != nil {
		return err
	}
	return nil
}

func (cc *CryptoConnection) decrypt() {
	for i := 0; i < len(cc.readBuffer); i += cc.cipher.BlockSize() {
		cc.cipher.Decrypt(cc.readBuffer[i:i+cc.cipher.BlockSize()], cc.readBuffer[i:i+cc.cipher.BlockSize()])
	}
	cc.setReadBufferSize()
}

func (cc *CryptoConnection) encrypt() error {
	var err error
	err = cc.setWriteBufferSize()
	if err != nil {
		return err
	}
	unaligned := len(cc.writeBuffer) % cc.cipher.BlockSize()
	if unaligned != 0 {
		alignSize := cc.cipher.BlockSize() - unaligned
		cc.writeBuffer = append(cc.writeBuffer, make([]byte, alignSize)...)
	}
	for i := 0; i < len(cc.writeBuffer); i += cc.cipher.BlockSize() {
		cc.cipher.Encrypt(cc.writeBuffer[i:i+cc.cipher.BlockSize()], cc.writeBuffer[i:i+cc.cipher.BlockSize()])
	}
	return nil
}

func (cc *CryptoConnection) flush() error {
	err := cc.conn.WriteBytes(cc.writeBuffer)
	if err != nil {
		return err
	}
	cc.resetWriteBuffer()
	return nil
}

func (cc *CryptoConnection) resetReadBuffer() {
	cc.readBuffer = nil
	cc.readOffset = 0
}

func (cc *CryptoConnection) resetWriteBuffer() {
	cc.writeBuffer = make([]byte, 8) // sizeof uint64
}
