package librlcom

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

var (
	ErrUnknownMessage = errors.New("unknown message")
)

type MessageType uint64

// Message types
const (
	MTUnknown MessageType = 0x00
	MTOK      MessageType = 0x0C << 56
	MTNotOK   MessageType = 0x7C << 56

	MTVersion     MessageType = 0x01
	MTToken       MessageType = 0x70
	MTLinkLibrary MessageType = 0x111B

	MTGetSymbolLibrary      MessageType = 0x63751711B
	MTResolvedSymbolLibrary MessageType = 0x435013D417
)

type Header struct {
	MessageType
}

type Message interface {
	io.ReaderFrom
	io.WriterTo
	Size() int64
}

type String string
type Empty struct{}

type Unknown struct{ Empty }
type OK struct{ Empty }
type NotOK struct{ Empty }

type Version struct{ Value uint64 }
type Token struct{ Token []byte }
type LinkLibrary struct{ String }

type GetSymbolLibrary struct{ String }
type ResolvedSymbolLibrary struct{ String }

var typeMapMsg2Go = map[MessageType]reflect.Type{
	MTUnknown: reflect.TypeOf(Unknown{}),
	MTOK:      reflect.TypeOf(OK{}),
	MTNotOK:   reflect.TypeOf(NotOK{}),

	MTVersion:     reflect.TypeOf(Version{}),
	MTToken:       reflect.TypeOf(Token{}),
	MTLinkLibrary: reflect.TypeOf(LinkLibrary{}),

	MTGetSymbolLibrary:      reflect.TypeOf(GetSymbolLibrary{}),
	MTResolvedSymbolLibrary: reflect.TypeOf(ResolvedSymbolLibrary{}),
}

var typeMapGo2Msg = map[reflect.Type]MessageType{
	reflect.TypeOf(&Unknown{}): MTUnknown,
	reflect.TypeOf(&OK{}):      MTOK,
	reflect.TypeOf(&NotOK{}):   MTNotOK,

	reflect.TypeOf(&Version{}):     MTVersion,
	reflect.TypeOf(&Token{}):       MTToken,
	reflect.TypeOf(&LinkLibrary{}): MTLinkLibrary,

	reflect.TypeOf(&GetSymbolLibrary{}):      MTGetSymbolLibrary,
	reflect.TypeOf(&ResolvedSymbolLibrary{}): MTResolvedSymbolLibrary,
}

func (h *Header) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, binary.LittleEndian, h)
	if err != nil {
		return 0, err
	}
	return h.Size(), nil
}

func (h *Header) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.LittleEndian, h)
	if err != nil {
		return 0, err
	}
	return h.Size(), nil
}

func (h *Header) Size() int64 {
	return int64(binary.Size(*h))
}

func (e *Empty) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, nil
}

func (e *Empty) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (e *Empty) Size() int64 {
	return 0
}

func (s *String) ReadFrom(r io.Reader) (n int64, err error) {
	var length uint64
	err = binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return 0, err
	}
	var m int
	str := make([]byte, length)
	for k := uint64(0); k < length; k += uint64(m) {
		m, err = r.Read(str[k:])
		if err != nil {
			*s = String(str)
			return 8 + int64(k), err
		}
	}
	*s = String(str)
	return 8 + int64(length), nil
}

func (s *String) WriteTo(w io.Writer) (n int64, err error) {
	var length uint64 = uint64(len(*s))
	err = binary.Write(w, binary.LittleEndian, &length)
	if err != nil {
		return 0, err
	}
	m, err := io.WriteString(w, string(*s))
	if err != nil {
		return 8 + int64(m), err
	}
	return 8 + int64(len(*s)), nil
}

func (s *String) Size() int64 {
	return int64(len(*s)) + 1
}

func (s *String) String() string {
	return string(*s)
}

func (v *Version) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, binary.LittleEndian, &v.Value)
	if err != nil {
		return 0, err
	}
	return v.Size(), nil
}

func (v *Version) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.LittleEndian, &v.Value)
	if err != nil {
		return 0, err
	}
	return v.Size(), nil
}

func (v *Version) Size() int64 {
	return int64(binary.Size(*v))
}

func (t *Token) ReadFrom(r io.Reader) (n int64, err error) {
	var length uint64
	err = binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return 0, err
	}
	var m int
	t.Token = make([]byte, length)
	for k := uint64(0); k < length; k += uint64(m) {
		m, err = r.Read(t.Token[k:])
		if err != nil {
			return 8 + int64(k), err
		}
	}
	return 8 + int64(length), nil
}

func (t *Token) WriteTo(w io.Writer) (n int64, err error) {
	var length uint64 = uint64(len(t.Token))
	err = binary.Write(w, binary.LittleEndian, &length)
	if err != nil {
		return 0, err
	}
	var m int
	for k := uint64(0); k < length; k += uint64(m) {
		m, err = w.Write(t.Token[k:])
		if err != nil {
			return 8 + int64(k), err
		}
	}
	return 8 + int64(len(t.Token)), nil
}

func (t *Token) Size() int64 {
	return int64(len(t.Token))
}
