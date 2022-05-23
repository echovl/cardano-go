package cardano

import (
	"testing"
)

func TestAddress(t *testing.T) {
	tests := []struct {
		address string
	}{
		{"addr_test1vqgjd0t02q9yglcjwdc8dht9tz6gkfpqqm7evs5csrklakcqmwv40"},
	}
	for _, tt := range tests {
		a := Address(tt.address)
		data := a.Bytes()
		_, got, _ := DecodeAddress(data)
		if string(got) != tt.address {
			t.Fatalf("got %v, want %v", got, tt.address)
		}
	}
}
