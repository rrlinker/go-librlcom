package librlcom

import (
	"encoding/binary"
	"io"
	"reflect"
)

type Type uint64

// Message types
const (
	MTOK Type = 0x0C << 56

	MTVersion       Type = 0x01
	MTAuthorization Type = 0x0A
	MTLinkLibrary   Type = 0x111B

	MTGetSymbolLibrary      Type = 0x63751711B
	MTResolvedSymbolLibrary Type = 0x435013D417
)

type Header struct {
	MessageSize uint64
	MessageType Type
}

type Message interface {
	io.ReaderFrom
	io.WriterTo
	Size() int64
}

type Token [128]byte
type String string

type OK struct{}

type Version struct{ Value uint64 }
type Authorization struct{ Token }
type LinkLibrary struct{ String }

type GetSymbolLibrary struct{ String }
type ResolvedSymbolLibrary struct{ String }

var typeMapMsg2Go = map[Type]reflect.Type{
	MTOK: reflect.TypeOf(OK{}),

	MTVersion:       reflect.TypeOf(Version{}),
	MTAuthorization: reflect.TypeOf(Authorization{}),
	MTLinkLibrary:   reflect.TypeOf(LinkLibrary{}),

	MTGetSymbolLibrary:      reflect.TypeOf(GetSymbolLibrary{}),
	MTResolvedSymbolLibrary: reflect.TypeOf(ResolvedSymbolLibrary{}),
}

var typeMapGo2Msg = map[reflect.Type]Type{
	reflect.TypeOf(&OK{}): MTOK,

	reflect.TypeOf(&Version{}):       MTVersion,
	reflect.TypeOf(&Authorization{}): MTAuthorization,
	reflect.TypeOf(&LinkLibrary{}):   MTLinkLibrary,

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

func (ok *OK) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, nil
}

func (ok *OK) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (ok *OK) Size() int64 {
	return 0
}

func (t *Token) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, binary.LittleEndian, (*[128]byte)(t))
	if err != nil {
		return 0, err
	}
	return t.Size(), nil
}

func (t *Token) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, binary.LittleEndian, (*[128]byte)(t))
	if err != nil {
		return 0, err
	}
	return t.Size(), nil
}

func (t *Token) Size() int64 {
	return int64(binary.Size(*t))
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
			return int64(k), err
		}
	}
	*s = String(str)
	return int64(length), nil
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
