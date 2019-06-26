package librlcom

import (
	"bytes"
	"net"
	"sync"
	"testing"
)

func TestConnection(t *testing.T) {
	var err error
	var wg sync.WaitGroup
	clientConn, serverConn := net.Pipe()

	client := NewConnection(clientConn)
	server := NewConnection(serverConn)

	defer client.Close()
	defer server.Close()

	wg.Add(2)

	want := []byte("The quick brown fox jumps over the lazy dog")
	go func() {
		defer wg.Done()
		got, err := server.ReadBytes()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want, got) {
			t.Fatalf("want = %v | got = %v\n", want, got)
		}
	}()
	err = client.WriteBytes(want)
	wg.Done()
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
