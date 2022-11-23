package wallet

import (
	"testing"

	"github.com/tyler-smith/go-bip39"
)

type TestVector struct {
	mnemonic     string
	rootXsk      string
	addrXsk0     string
	addrXvk0     string
	addrXsk1     string
	addrXvk1     string
	paymentAddr0 string
	paymentAddr1 string
	stakeAddr    string
}

var testVectors = []TestVector{
	// 15 words
	{
		mnemonic:     "art forum devote street sure rather head chuckle guard poverty release quote oak craft enemy",
		rootXsk:      "root_xsk1hretan5mml3tq2p0twkhq4tz4jvka7m2l94kfr6yghkyfar6m9wppc7h9unw6p65y23kakzct3695rs32z7vaw3r2lg9scmfj8ec5du3ufydu5yuquxcz24jlkjhsc9vsa4ufzge9s00fn398svhacse5sh85djs",
		addrXsk0:     "addr_xsk1fzgcl9km0mve2jwe8qxve364w6te9vpddhwpw5g8wnjlupmmm9wxpdda6jaglx7smwl6qd5xuzjcweeq8ykp0wg9hng4pg6eumwx2t90swaed7ehsa6j86qsw3fnl4thtemsng6vukmz6ddf3cnd4sfkzu74xjqg",
		addrXvk0:     "addr_xvk1fz009r4f0aceaemksezlca9cz8p8rewhaurvyvgg2ndnq9vwj3w6lqamjman0pm4y05pqazn8l2hwhnhpx35eedk9566nr3xmtqnv9ccm4zyu",
		addrXsk1:     "addr_xsk1lq2ylz7fhsn0dfmul2pe833cdwvjnvux9uaxuzaz50gs7pnmm9wq343uh5cpfs87tgh9saa86un8e2l266rsge0c5qsmtaud5r64ctndwkyth8q07fgusyr3fldhn6lgd5tat5cmcdzvfzhtd0cpsleuxg3sakhv",
		addrXvk1:     "addr_xvk1y3r70ejyadsaplez83p7uhy8p6l08a5sjl860kszevxu0jaxcwmx6avghwwqluj3eqg8zn7m0847smgh6hf3hs6ycj9wk6lsrplncvsxqj6wd",
		paymentAddr0: "addr_test1vpu5vlrf4xkxv2qpwngf6cjhtw542ayty80v8dyr49rf5eg57c2qv",
		paymentAddr1: "addr_test1vq0a2lgc2e0r597dr983jrf5ns4hxz027u8n7wlcsjcw4ks96yjys",
		stakeAddr:    "stake_test1urxr8x34l8s0uquu75gvwcw5m55sgrzga9jhlkk8a8qpm9q9p2w0s",
	},
	// 18 words
	{
		mnemonic:     "churn shaft spoon second erode useless thrive burst group seed element sign scrub buffalo jelly grace neck useless",
		rootXsk:      "root_xsk1az4qjp85qunj75m8krdvdygmv6u4ceqj8vnwaf38wfd69ycksa0fwt7n0cfp5zwmht9u0j9dzxxnfssjmkh4vn3dwxvddsle6m2vkm8q8p7addwq8y7q3s3eekd3ate40rfr6rpjakctcn2p54cpr3kjmy2hdej2",
		addrXsk0:     "addr_xsk17zj2lhjk379klp40xfzsad0yzygqe45uaggnkzf4ld3emgsksa00ftq88fnfjxg245kjjqcukyjfg4lwmf3r2qqymyyqennch3y8llyg0d629pdx0pp0l69lerjz75kxmk5e6cr2d82kafp7a25y0qy5fvj06w0u",
		addrXvk0:     "addr_xvk1fwgdh5vv6akdc3rjpeq57xxq4lc9m84xcrt6q827mq7u20wuw54gs7m552z6v7zzll5tlj8y9afvdhdfn4sx56w4d6jra64gg7qfgjc8e8sau",
		addrXsk1:     "addr_xsk1wz99hznmt96crxthcmnxqttaul6caq4hv5jwttd5lly2mfsksa08q68skn2ggclu6vf40phx3wnj4e8fvxed6at8xxekwa49rg4c3ec8kp2nwcxfw6sgxphzckg5v0dausldvya0w6jy5k3cxwrqdjsthqls7zxn",
		addrXvk1:     "addr_xvk135hqmkaqydnxnq6wmjkkhasvwjprpnqnzsrwwes6mql45enlcsqs0vz4xasvja4qsvrw93v3gc7mmep76cf67a4yffdrsvuxqm9qhwqlvnay5",
		paymentAddr0: "addr_test1vptvyjfjvs7wdn583rv3th3fvf9fauv5f6gylkhh5k245zcdjvdac",
		paymentAddr1: "addr_test1vr3nq3kyg9c9t4nn6a5zymz3at3zsmcr9lkqxghxh5v822gcu7ava",
		stakeAddr:    "stake_test1uzx24u6vjgmgfzqg38uqrmrs720pum2sgxqntv0yzpvf72qhlmveq",
	},
	// 21 words
	{
		mnemonic:     "draft ability female child jump maid roof hurt below live topple paper exclude ordinary coach churn sunset emerge blame ketchup much",
		rootXsk:      "root_xsk17zqw352yj02seytp9apunec55722k93crtplq8chgpfh7cx33dg4v2x3wpyhd9chhkknzhprztumrystkpfl5nyhyeuq0gnwf76r39u9l9q3z40hgf5jv6xn8unr5acs3yy8fxg35v5xjsw4kwvf5zfkvcn57fle",
		addrXsk0:     "addr_xsk1fz8tz0pdda8la0aqhadnzctw0p48zwygkgf4xyar2jjljm733dgkprs4sj8cxfwv9xtfddpdfvjlap0hhg9gd37pr0tp7ue48mh9cnfyy68k52f88z5vghezam30c3pcue6aewl4mqul6nvassxlenh3eqh822k2",
		addrXvk0:     "addr_xvk1x4dme9s2f5xxn77wgjhggqh73r6syy4nvjcdjklnaqrh48f6desjgf50dg5jww9gc30j9mhzl3zr3en4mjaltkpel4xempqdln80rjq5grmc7",
		addrXsk1:     "addr_xsk18peu0v64maghaa87jvu0txdkftvznq7he2yhntk8eem56mk33dgl2rwt8kmhcgdytr6fjn0t4cdf6sr3xud67yhwnjhzyghgu294f6v0fcfzqlactzd8cf5m4tpu7yyn5x58dx6q00d362j6e06g88phjglns3p5",
		addrXvk1:     "addr_xvk1ndtepmpg06x9nskfasvr50mue356e4rqlvuzf8jjcj6n48feexsg7nsjyplmsky60snfh2kreugf8gdgw6d5q77mr5494jl5swwr0ysprauul",
		paymentAddr0: "addr_test1vz83dnlqqtdrlct4kz3f7d07d59w6p4yrtlr62340yklhaqrrykc7",
		paymentAddr1: "addr_test1vzr08acccp7s3l9cppvptz7jyflejkkuma2k06vx4vjrcqsl4gkk5",
		stakeAddr:    "stake_test1urvx673x97g9dadcs2hjse5luwpsempg5fny6u56lx7vv3gpzreln",
	},
}

func TestCreateWallet(t *testing.T) {
	for _, testVector := range testVectors {
		client := NewClient(&Options{})
		defer client.Close()

		newEntropy = func(bitSize int) []byte {
			entropy, err := bip39.EntropyFromMnemonic(testVector.mnemonic)
			if err != nil {
				t.Error(err)
			}
			return entropy
		}

		w, mnemonic, err := client.CreateWallet("test", "")
		if err != nil {
			t.Error(err)
		}

		addrXsk0 := bech32From("addr_xsk", w.addrKeys[0])
		addrXvk0 := bech32From("addr_xvk", w.addrKeys[0].XPubKey())

		if addrXsk0 != testVector.addrXsk0 {
			t.Errorf("invalid addrXsk0 :\ngot: %v\nwant: %v", addrXsk0, testVector.addrXsk0)
		}

		if addrXvk0 != testVector.addrXvk0 {
			t.Errorf("invalid addrXvk0 :\ngot: %v\nwant: %v", addrXvk0, testVector.addrXvk0)
		}

		addresses, err := w.Addresses()
		if err != nil {
			t.Fatal(err)
		}

		if mnemonic != testVector.mnemonic {
			t.Errorf("invalid mnemonic:\ngot: %v\nwant: %v", mnemonic, testVector.mnemonic)
		}

		if addresses[0].Bech32() != testVector.paymentAddr0 {
			t.Errorf("invalid paymentAddr0:\ngot: %v\nwant: %v", addresses[0], testVector.paymentAddr0)
		}

		stakeAddr, err := w.StakeAddress()
		if err != nil {
			t.Fatal(err)
		}

		if stakeAddr.Bech32() != testVector.stakeAddr {
			t.Errorf("invalid stakeAddr:\ngot: %v\nwant: %v", stakeAddr, testVector.stakeAddr)
		}
	}
}

func TestRestoreWallet(t *testing.T) {
	for _, testVector := range testVectors {
		client := NewClient(&Options{})
		defer client.Close()

		w, err := client.RestoreWallet("test", "", testVector.mnemonic)
		if err != nil {
			t.Error(err)
		}

		addrXsk0 := bech32From("addr_xsk", w.addrKeys[0])
		addrXvk0 := bech32From("addr_xvk", w.addrKeys[0].XPubKey())

		if addrXsk0 != testVector.addrXsk0 {
			t.Errorf("invalid addrXsk0 :\ngot: %v\nwant: %v", addrXsk0, testVector.addrXsk0)
		}

		if addrXvk0 != testVector.addrXvk0 {
			t.Errorf("invalid addrXvk0 :\ngot: %v\nwant: %v", addrXvk0, testVector.addrXvk0)
		}

		addresses, err := w.Addresses()
		if err != nil {
			t.Fatal(err)
		}

		if addresses[0].Bech32() != testVector.paymentAddr0 {
			t.Errorf("invalid paymentAddr0:\ngot: %v\nwant: %v", addresses[0], testVector.paymentAddr0)
		}

		stakeAddr, err := w.StakeAddress()
		if err != nil {
			t.Fatal(err)
		}

		if stakeAddr.Bech32() != testVector.stakeAddr {
			t.Errorf("invalid stakeAddr:\ngot: %v\nwant: %v", stakeAddr, testVector.stakeAddr)
		}
	}
}
