package bech32

import "testing"

func TestDecode(t *testing.T) {
	decoded := "root_xsk1spewah0xnqc5lk4hxcpunlmyv76naz69kwm9nvpjzal8fepwye8mu222f0pq39wam0mqf3wgk28xvjl3rn0fheuql272wp2qlgu88tmzjmjfckn90lf52l8cfysy66k53dt2dzqjusmzmkk7tfltq4grku60xxg2"

	hrp, _, _ := Decode(decoded)

	if hrp != "root_xsk" {
		t.Errorf("invalid decode output:\ngot: %v\nwant: %v", hrp, "root_xsk")
	}
}
