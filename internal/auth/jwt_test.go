package auth

import (
	"testing"
	"time"
)

func TestSignRoundTrip(t *testing.T) {
	tok, err := SignAccessToken("test-secret", "user-1", "a@b.com", "free", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseAccessToken("test-secret", tok)
	if err != nil {
		t.Fatal(err)
	}
	if UserID(c) != "user-1" {
		t.Fatalf("subject %q", UserID(c))
	}
	if c.Email != "a@b.com" || c.Plan != "free" {
		t.Fatalf("claims %+v", c)
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := ParseAccessToken("secret", "not-a-jwt")
	if err == nil {
		t.Fatal("expected error")
	}
}
