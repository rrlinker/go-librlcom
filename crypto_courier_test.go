package librlcom

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"sync"
	"testing"
)

func TestCryptoCourierWithKey(t *testing.T) {
	var err error
	var wg sync.WaitGroup
	clientConn, serverConn := net.Pipe()

	client := NewCryptoCourier(clientConn)
	server := NewCryptoCourier(serverConn)

	defer client.Close()
	defer server.Close()

	key := make([]byte, 16)
	_, err = rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(2)

	want := LinkLibrary{"library"}
	go func() {
		defer wg.Done()
		err := server.InitWithKey(key)
		if err != nil {
			t.Fatal(err)
		}
		rawGot, err := server.Receive()
		if err != nil {
			t.Log(rawGot)
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
	err = client.InitWithKey(key)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(&want)
	wg.Done()
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}

func TestCryptoCourier(t *testing.T) {
	var err error
	var wg sync.WaitGroup
	clientConn, serverConn := net.Pipe()

	client := NewCryptoCourier(clientConn)
	server := NewCryptoCourier(serverConn)

	defer client.Close()
	defer server.Close()

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(2)

	want := LinkLibrary{"library"}
	go func() {
		defer wg.Done()
		err := server.InitAsServer(priv)
		if err != nil {
			t.Fatal(err)
		}
		rawGot, err := server.Receive()
		if err != nil {
			t.Log(rawGot)
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
	err = client.InitAsClient(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send(&want)
	wg.Done()
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
