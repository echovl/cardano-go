package bech32

import (
	"encoding/hex"
	"testing"

	"github.com/echovl/cardano-go/bech32/prefixes"
)

var (
	tests = []struct {
		be32        string
		prefix      string
		dataB32     []byte // "byte" encoded in 5 bits, max value per-byte 32
		dataB256    []byte // "byte" encoded in 8 bits, max value per-byte 256
		dataB256Hex string
	}{
		{
			be32:        "addr_vk1w0l2sr2zgfm26ztc6nl9xy8ghsk5sh6ldwemlpmp9xylzy4dtf7st80zhd",
			prefix:      prefixes.AddrPublicKey,
			dataB32:     []byte{0xe, 0xf, 0x1f, 0xa, 0x10, 0x3, 0xa, 0x2, 0x8, 0x9, 0x1b, 0xa, 0x1a, 0x2, 0xb, 0x18, 0x1a, 0x13, 0x1f, 0x5, 0x6, 0x4, 0x7, 0x8, 0x17, 0x10, 0x16, 0x14, 0x10, 0x17, 0x1a, 0x1f, 0xd, 0xe, 0x19, 0x1b, 0x1f, 0x1, 0x1b, 0x1, 0x5, 0x6, 0x4, 0x1f, 0x2, 0x4, 0x15, 0xd, 0xb, 0x9, 0x1e, 0x10},
			dataB256:    []byte{0x73, 0xfe, 0xa8, 0xd, 0x42, 0x42, 0x76, 0xad, 0x9, 0x78, 0xd4, 0xfe, 0x53, 0x10, 0xe8, 0xbc, 0x2d, 0x48, 0x5f, 0x5f, 0x6b, 0xb3, 0xbf, 0x87, 0x61, 0x29, 0x89, 0xf1, 0x12, 0xad, 0x5a, 0x7d},
			dataB256Hex: "73fea80d424276ad0978d4fe5310e8bc2d485f5f6bb3bf87612989f112ad5a7d",
		},
		{
			be32:        "stake_vk1px4j0r2fk7ux5p23shz8f3y5y2qam7s954rgf3lg5merqcj6aetsft99wu",
			prefix:      prefixes.StakePublicKey,
			dataB32:     []byte{0x1, 0x6, 0x15, 0x12, 0xf, 0x3, 0xa, 0x9, 0x16, 0x1e, 0x1c, 0x6, 0x14, 0x1, 0xa, 0x11, 0x10, 0x17, 0x2, 0x7, 0x9, 0x11, 0x4, 0x14, 0x4, 0xa, 0x0, 0x1d, 0x1b, 0x1e, 0x10, 0x5, 0x14, 0x15, 0x3, 0x8, 0x9, 0x11, 0x1f, 0x8, 0x14, 0x1b, 0x19, 0x3, 0x0, 0x18, 0x12, 0x1a, 0x1d, 0x19, 0xb, 0x10},
			dataB256:    []byte{0x9, 0xab, 0x27, 0x8d, 0x49, 0xb7, 0xb8, 0x6a, 0x5, 0x51, 0x85, 0xc4, 0x74, 0xc4, 0x94, 0x22, 0x81, 0xdd, 0xfa, 0x5, 0xa5, 0x46, 0x84, 0xc7, 0xe8, 0xa6, 0xf2, 0x30, 0x62, 0x5a, 0xee, 0x57},
			dataB256Hex: "09ab278d49b7b86a055185c474c4942281ddfa05a54684c7e8a6f230625aee57",
		},
		{
			be32:        "script1cda3khwqv60360rp5m7akt50m6ttapacs8rqhn5w342z7r35m37",
			prefix:      prefixes.Script,
			dataB32:     []byte{0x18, 0xd, 0x1d, 0x11, 0x16, 0x17, 0xe, 0x0, 0xc, 0x1a, 0xf, 0x11, 0x1a, 0xf, 0x3, 0x1, 0x14, 0x1b, 0x1e, 0x1d, 0x16, 0xb, 0x14, 0xf, 0x1b, 0x1a, 0xb, 0xb, 0x1d, 0x1, 0x1d, 0x18, 0x10, 0x7, 0x3, 0x0, 0x17, 0x13, 0x14, 0xe, 0x11, 0x15, 0xa, 0x2, 0x1e},
			dataB256:    []byte{0xc3, 0x7b, 0x1b, 0x5d, 0xc0, 0x66, 0x9f, 0x1d, 0x3c, 0x61, 0xa6, 0xfd, 0xdb, 0x2e, 0x8f, 0xde, 0x96, 0xbe, 0x87, 0xb8, 0x81, 0xc6, 0xb, 0xce, 0x8e, 0x8d, 0x54, 0x2f},
			dataB256Hex: "c37b1b5dc0669f1d3c61a6fddb2e8fde96be87b881c60bce8e8d542f",
		},
		{
			be32:        "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x",
			prefix:      prefixes.Addr,
			dataB32:     []byte{0x0, 0x6, 0xa, 0x9, 0x6, 0xc, 0xa, 0x1c, 0x1b, 0x4, 0x17, 0xb, 0xb, 0x16, 0x6, 0x4, 0x6, 0x1, 0x7, 0x6, 0xf, 0xd, 0x1f, 0x1, 0xd, 0xb, 0x11, 0x16, 0x1a, 0x18, 0xe, 0x13, 0x8, 0x14, 0x1, 0x6, 0x12, 0x11, 0x12, 0x17, 0x10, 0x4, 0xd, 0x2, 0x19, 0x3, 0x11, 0x13, 0xf, 0xd, 0x11, 0xc, 0x1f, 0x1f, 0x1b, 0x4, 0x0, 0xe, 0x10, 0x6, 0x14, 0xe, 0x16, 0xb, 0x18, 0xd, 0x7, 0x18, 0x18, 0x11, 0x10, 0x0, 0x7, 0x11, 0x14, 0x1f, 0x1c, 0x1e, 0xd, 0x3, 0xc, 0xa, 0x6, 0xe, 0x1f, 0xa, 0xe, 0x4, 0xe, 0x9, 0x8, 0x10},
			dataB256:    []byte{0x1, 0x94, 0x93, 0x31, 0x5c, 0xd9, 0x2e, 0xb5, 0xd8, 0xc4, 0x30, 0x4e, 0x67, 0xb7, 0xe1, 0x6a, 0xe3, 0x6d, 0x61, 0xd3, 0x45, 0x2, 0x69, 0x46, 0x57, 0x81, 0x1a, 0x2c, 0x8e, 0x33, 0x7b, 0x62, 0xcf, 0xff, 0x64, 0x3, 0xa0, 0x6a, 0x3a, 0xcb, 0xc3, 0x4f, 0x8c, 0x46, 0x0, 0x3c, 0x69, 0xfe, 0x79, 0xa3, 0x62, 0x8c, 0xef, 0xa9, 0xc4, 0x72, 0x51},
			dataB256Hex: "019493315cd92eb5d8c4304e67b7e16ae36d61d34502694657811a2c8e337b62cfff6403a06a3acbc34f8c46003c69fe79a3628cefa9c47251",
		},
		{
			be32:        "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k",
			prefix:      prefixes.Addr,
			dataB32:     []byte{0x8, 0x6, 0xa, 0x9, 0x6, 0xc, 0xa, 0x1c, 0x1b, 0x4, 0x17, 0xb, 0xb, 0x16, 0x6, 0x4, 0x6, 0x1, 0x7, 0x6, 0xf, 0xd, 0x1f, 0x1, 0xd, 0xb, 0x11, 0x16, 0x1a, 0x18, 0xe, 0x13, 0x8, 0x14, 0x1, 0x6, 0x12, 0x11, 0x12, 0x17, 0x10, 0x4, 0xd, 0x2, 0x19, 0x3, 0x14, 0x1, 0x13, 0x2, 0x1e, 0x14, 0x6, 0x6, 0x18, 0x3},
			dataB256:    []byte{0x41, 0x94, 0x93, 0x31, 0x5c, 0xd9, 0x2e, 0xb5, 0xd8, 0xc4, 0x30, 0x4e, 0x67, 0xb7, 0xe1, 0x6a, 0xe3, 0x6d, 0x61, 0xd3, 0x45, 0x2, 0x69, 0x46, 0x57, 0x81, 0x1a, 0x2c, 0x8e, 0x81, 0x98, 0xbd, 0x43, 0x1b, 0x3},
			dataB256Hex: "419493315cd92eb5d8c4304e67b7e16ae36d61d34502694657811a2c8e8198bd431b03",
		},
		{
			be32:        "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8",
			prefix:      prefixes.Addr,
			dataB32:     []byte{0xc, 0x6, 0xa, 0x9, 0x6, 0xc, 0xa, 0x1c, 0x1b, 0x4, 0x17, 0xb, 0xb, 0x16, 0x6, 0x4, 0x6, 0x1, 0x7, 0x6, 0xf, 0xd, 0x1f, 0x1, 0xd, 0xb, 0x11, 0x16, 0x1a, 0x18, 0xe, 0x13, 0x8, 0x14, 0x1, 0x6, 0x12, 0x11, 0x12, 0x17, 0x10, 0x4, 0xd, 0x2, 0x19, 0x3, 0x10},
			dataB256:    []byte{0x61, 0x94, 0x93, 0x31, 0x5c, 0xd9, 0x2e, 0xb5, 0xd8, 0xc4, 0x30, 0x4e, 0x67, 0xb7, 0xe1, 0x6a, 0xe3, 0x6d, 0x61, 0xd3, 0x45, 0x2, 0x69, 0x46, 0x57, 0x81, 0x1a, 0x2c, 0x8e},
			dataB256Hex: "619493315cd92eb5d8c4304e67b7e16ae36d61d34502694657811a2c8e",
		},
	}
)

// test encoding from string
func TestEncodeWithHrpAndBytes(t *testing.T) {
	for _, test := range tests {
		hrp, data, err := DecodeNoLimit(test.be32)
		if err != nil {
			t.Fatal(err)
		}
		if hrp != test.prefix {
			t.Errorf("decoded wrong prefix: %s expected %s", hrp, test.prefix)
		}

		if hex.EncodeToString(data) != hex.EncodeToString(test.dataB32) {
			t.Errorf("decoded wrong data: %#v expected %#v", data, test.dataB32)
		}

		be32, err := Encode(test.prefix, data)
		if err != nil {
			t.Fatal(err)
		}
		if be32 != test.be32 {
			t.Errorf("encoded wrong: %s expected %s", be32, test.be32)
		}
	}
}

// test encoding from bytes
func TestEncodeFromBase256WithHrpAndBytes(t *testing.T) {
	for _, test := range tests {
		hrp, data, err := DecodeToBase256(test.be32)
		if err != nil {
			t.Fatal(err)
		}
		if hrp != test.prefix {
			t.Errorf("decoded wrong prefix: %s expected %s", hrp, test.prefix)
		}

		if hex.EncodeToString(data) != test.dataB256Hex {
			t.Errorf("decoded wrong data: %#v expected %#v", data, test.dataB256Hex)
		}

		be32, err := EncodeFromBase256(test.prefix, data)
		if err != nil {
			t.Fatal(err)
		}
		if be32 != test.be32 {
			t.Errorf("encoded wrong: %s expected %s", be32, test.be32)
		}
	}
}

// test Codec implementations
type AddrPublicKey [32]byte

func (vk AddrPublicKey) Bytes() []byte      { return vk[:] }
func (vk AddrPublicKey) Len() int           { return len(vk) }
func (vk *AddrPublicKey) SetBytes(b []byte) { copy((*vk)[:], b) }
func (vk AddrPublicKey) Prefix() string     { return prefixes.AddrPublicKey }

type StakePublicKey struct {
	AddrPublicKey
}

func (vk StakePublicKey) Prefix() string { return prefixes.StakePublicKey }

func TestBech32Codec(t *testing.T) {
	want1 := tests[0]
	want2 := tests[1]
	got1 := AddrPublicKey{}
	got2 := StakePublicKey{}
	if err := DecodeInto(want1.be32, &got1); err != nil {
		t.Fatal(err)
	}
	if err := DecodeInto(want2.be32, &got2); err != nil {
		t.Fatal(err)
	}

	if got1.Prefix() != want1.prefix {
		t.Errorf("decoded wrong prefix: %s expected %s", got1.Prefix(), want1.prefix)
	}

	if hex.EncodeToString(got1.Bytes()) != want1.dataB256Hex {
		t.Errorf("decoded wrong data: %#v expected %#v", got1.Bytes(), want1.dataB256Hex)
	}

	if got2.Prefix() != want2.prefix {
		t.Errorf("decoded wrong prefix: %s expected %s", got1.Prefix(), want1.prefix)
	}

	if hex.EncodeToString(got2.Bytes()) != want2.dataB256Hex {
		t.Errorf("decoded wrong data: %#v expected %#v", got2.Bytes(), want2.dataB256Hex)
	}

	if s, err := Encode(got1); err != nil {
		t.Fatal(err)
	} else {
		if s != want1.be32 {
			t.Errorf("encoded wrong Bech32Coded: %s expected %s", s, want1.be32)
		}
	}

	if s, err := EncodeFromBase256(got1); err != nil {
		t.Fatal(err)
	} else {
		if s != want1.be32 {
			t.Errorf("encoded wrong Bech32Coded: %s expected %s", s, want1.be32)
		}
	}

	if s, err := Encode(got2); err != nil {
		t.Fatal(err)
	} else {
		if s != want2.be32 {
			t.Errorf("encoded wrong Bech32Coded: %s expected %s", s, want2.be32)
		}
	}

	if s, err := EncodeFromBase256(got2); err != nil {
		t.Fatal(err)
	} else {
		if s != want2.be32 {
			t.Errorf("encoded wrong Bech32Coded: %s expected %s", s, want2.be32)
		}
	}

}
