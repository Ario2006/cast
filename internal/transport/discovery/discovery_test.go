package discovery

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestRandomCodeAndValidation(t *testing.T) {
	code, err := RandomCode(6)
	if err != nil {
		t.Fatalf("RandomCode error = %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("RandomCode length = %d, want 6", len(code))
	}
	if !ValidCode(code) {
		t.Fatalf("ValidCode(%q) = false, want true", code)
	}
	if ValidCode("abc") { // too short
		t.Fatal("ValidCode(abc) = true, want false")
	}
	if ValidCode("TOOLONGCODE") { // too long
		t.Fatal("ValidCode(TOOLONGCODE) = true, want false")
	}
	if ValidCode("1234") { // contains '1' which is excluded from alphabet
		t.Fatal("ValidCode(1234) = true, want false")
	}
}

func TestRandomToken(t *testing.T) {
	token1, err := RandomToken()
	if err != nil {
		t.Fatal(err)
	}
	token2, err := RandomToken()
	if err != nil {
		t.Fatal(err)
	}
	if token1 == token2 || len(token1) < 32 {
		t.Fatalf("tokens: %q, %q", token1, token2)
	}
}

func TestLocalIPv4(t *testing.T) {
	ip := LocalIPv4()
	if ip == nil {
		t.Fatal("LocalIPv4 returned nil")
	}
}

func TestBroadcastQueryAndResponder(t *testing.T) {
	testPort := 39499
	responder, err := StartResponder(testPort, func(query []byte, sender *net.UDPAddr) ([]byte, bool) {
		if string(query) == "ping" {
			return []byte("pong"), true
		}
		return nil, false
	})
	if err != nil {
		t.Skipf("bind test UDP port %d: %v", testPort, err)
	}
	defer responder.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var gotResponse string
	err = BroadcastQuery(ctx, testPort, []byte("ping"), func(data []byte) bool {
		gotResponse = string(data)
		return true
	})
	if err != nil {
		t.Fatalf("BroadcastQuery error = %v", err)
	}
	if gotResponse != "pong" {
		t.Fatalf("gotResponse = %q, want pong", gotResponse)
	}
}
