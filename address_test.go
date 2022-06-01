package cardano

import (
	"testing"

	"github.com/echovl/bech32"
	"github.com/echovl/cardano-go/crypto"
)

const (
	paymentKey = "addr_vk1w0l2sr2zgfm26ztc6nl9xy8ghsk5sh6ldwemlpmp9xylzy4dtf7st80zhd"
	stakeKey   = "stake_vk1px4j0r2fk7ux5p23shz8f3y5y2qam7s954rgf3lg5merqcj6aetsft99wu"
	scriptHash = "script1cda3khwqv60360rp5m7akt50m6ttapacs8rqhn5w342z7r35m37"
	addrType0  = "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3n0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgse35a3x"
	addrType1  = "addr1z8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gten0d3vllmyqwsx5wktcd8cc3sq835lu7drv2xwl2wywfgs9yc0hh"
	addrType2  = "addr1yx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzerkr0vd4msrxnuwnccdxlhdjar77j6lg0wypcc9uar5d2shs2z78ve"
	addrType3  = "addr1x8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gt7r0vd4msrxnuwnccdxlhdjar77j6lg0wypcc9uar5d2shskhj42g"
	addrType4  = "addr1gx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer5pnz75xxcrzqf96k"
	addrType5  = "addr128phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtupnz75xxcrtw79hu"
	addrType6  = "addr1vx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzers66hrl8"
	addrType7  = "addr1w8phkx6acpnf78fuvxn0mkew3l0fd058hzquvz7w36x4gtcyjy7wx"
)

func TestAddress(t *testing.T) {
	pvk, err := crypto.NewPubKey(paymentKey)
	if err != nil {
		t.Fatal(err)
	}
	svk, err := crypto.NewPubKey(stakeKey)
	if err != nil {
		t.Fatal(err)
	}
	_, script, err := bech32.DecodeToBase256(scriptHash)
	if err != nil {
		t.Fatal(err)
	}

	paymentAddrCred, err := NewAddrKeyCredential(pvk)
	if err != nil {
		t.Fatal(err)
	}
	stakeAddrCred, err := NewAddrKeyCredential(svk)
	if err != nil {
		t.Fatal(err)
	}
	scriptCred := StakeCredential{Type: ScriptCredential, ScriptHash: script}

	base0, err := NewBaseAddress(Mainnet, paymentAddrCred, stakeAddrCred)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := base0.Bech32(), addrType0; got != want {
		t.Errorf("invalid base address\ngot: %s\nwant: %s", got, want)
	}

	base1, err := NewBaseAddress(Mainnet, scriptCred, stakeAddrCred)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := base1.Bech32(), addrType1; got != want {
		t.Errorf("invalid base address\ngot: %s\nwant: %s", got, want)
	}

	base2, err := NewBaseAddress(Mainnet, paymentAddrCred, scriptCred)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := base2.Bech32(), addrType2; got != want {
		t.Errorf("invalid base address\ngot: %s\nwant: %s", got, want)
	}

	base3, err := NewBaseAddress(Mainnet, scriptCred, scriptCred)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := base3.Bech32(), addrType3; got != want {
		t.Errorf("invalid base address\ngot: %s\nwant: %s", got, want)
	}

	pointer0, err := NewPointerAddress(Mainnet, paymentAddrCred, Pointer{2498243, 27, 3})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := pointer0.Bech32(), addrType4; got != want {
		t.Errorf("invalid pointer address\ngot: %s\nwant: %s", got, want)
	}

	pointer1, err := NewPointerAddress(Mainnet, scriptCred, Pointer{2498243, 27, 3})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := pointer1.Bech32(), addrType5; got != want {
		t.Errorf("invalid pointer address\ngot: %s\nwant: %s", got, want)
	}

	enterprise0, err := NewEnterpriseAddress(Mainnet, paymentAddrCred)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := enterprise0.Bech32(), addrType6; got != want {
		t.Errorf("invalid enterprise address\ngot: %s\nwant: %s", got, want)
	}

	enterprise1, err := NewEnterpriseAddress(Mainnet, scriptCred)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := enterprise1.Bech32(), addrType7; got != want {
		t.Errorf("invalid enterprise address\ngot: %s\nwant: %s", got, want)
	}

}

func TestNat(t *testing.T) {
	testcases := []uint64{0, 127, 128, 255, 256275757658493284}

	for _, tc := range testcases {
		want := tc
		bytes := encodeToNat(want)
		got, _, err := decodeFromNat(bytes)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("invalid nat decoding\ngot: %v\nwant: %v", got, want)
		}

	}
}
