package librlcom

import (
	"net"
	"sync"
	"testing"
)

func TestRawCourier(t *testing.T) {
	var err error
	var wg sync.WaitGroup
	clientConn, serverConn := net.Pipe()

	client := NewRawCourier(clientConn)
	server := NewRawCourier(serverConn)

	defer client.Close()
	defer server.Close()

	wg.Add(2)

	want := LinkLibrary{"library"}
	go func() {
		defer wg.Done()
		rawGot, err := server.Receive()
		if err != nil {
			t.Fatal(err)
		}
		if got, ok := rawGot.(*LinkLibrary); ok {
			if want.String.String() != got.String.String() {
				t.Fatalf("want = %s | got = %s\n", want.String.String(), got.String.String())
			}
		} else {
			t.Fatal("got not LinkLibrary")
		}
	}()
	err = client.Send(&want)
	wg.Done()
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
