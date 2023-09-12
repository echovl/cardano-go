package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/blockfrost"
	"github.com/echovl/cardano-go/crypto"
	"github.com/tyler-smith/go-bip39"
	"log"
	"strings"
)

const (
	enterpriseType uint = iota
	baseType
)

var mnemFlag string
var addrTypeFlag uint
var networkFlag string
var receiverAddr string

func init() {
	mnemonicDefault := ""

	flag.StringVar(&mnemFlag, "mnemonic", mnemonicDefault, "Mnemonic to restore wallet")
	flag.UintVar(&addrTypeFlag, "type", baseType, "Enum of address type(0: Enterprise Address, 1: Base Address)")
	flag.StringVar(&networkFlag, "network", "mainnet", "The network ie mainnet or testnet")
	flag.StringVar(&receiverAddr, "sendTo", "Ae2tdPwUPEZ9WyNeyrKZJhXtFQ9UKmraSAXAUhzmChyesRV7J4cLHkKQfZi", "The address to send ada to")
	flag.Parse()

	if mnemFlag == "" {
		log.Fatal("mnemonic cannot be empty")
	}

	if len(strings.Split(mnemFlag, " ")) != 24 {
		log.Fatal("mnemonic should be 24 words long")
	}

}

func harden(num uint) uint32 {
	return uint32(0x80000000 + num)
}

func generateBaseAddress(net cardano.Network, entropy []byte) (addr cardano.Address, utxoKey crypto.XPrvKey, err error) {
	rootKey := crypto.FromBip39Entropy(
		entropy,
		[]byte{},
	)

	accountKey := rootKey.Derive(harden(1852)).Derive(harden(1815)).Derive(harden(0))

	utxoKey = accountKey.Derive(0).Derive(0)
	utxoPubKey := utxoKey.XPubKey()
	utxoPubKeyHash := utxoPubKey.PubKey()
	fmt.Println("utxoPubKeyHash: ", hex.EncodeToString(utxoPubKeyHash))

	stakePubKey := accountKey.Derive(2).Derive(0).XPubKey()
	stakeKeyHash := stakePubKey.PubKey()

	paymentKey, _ := cardano.NewKeyCredential(utxoPubKeyHash)
	stakeKey, _ := cardano.NewKeyCredential(stakeKeyHash)
	_ = stakeKey
	addr, _ = cardano.NewBaseAddress(
		net,
		paymentKey,
		//paymentKey,
		stakeKey,
	)
	return
}

func main() {
	net := cardano.Mainnet
	cli := blockfrost.NewNode(
		net,
		//os.Getenv("BLOCKFROST_PROJECT_ID"),
		"",
	)

	// get protocol parameters for linearfee formula
	// fee = TxFeeFixed + TxFeePerByte * byteCount
	pr, err := cli.ProtocolParams()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("CoinsPerUTXOWord: ", pr.CoinsPerUTXOWord)

	entropy, _ := bip39.EntropyFromMnemonic(mnemFlag)

	// Generate a base address on testnet from the rootkey
	// the utxoPrvKey is used to sign the transaction
	sourceAddr, utxoPrvKey, err := generateBaseAddress(net, entropy)
	//sourceAddr.BytronAddr = sourceAddr1

	fmt.Println("sourceAddr:", sourceAddr)
	if err != nil {
		log.Fatal(err)
	}

	// If no sendTo address is provided use the source address
	if receiverAddr == "" {
		receiverAddr = sourceAddr.String()
	}

	fmt.Println("sourceAddr pub: ", hex.EncodeToString(sourceAddr.Bytes()))
	//receiverAddr = sourceAddr.String()
	receiver, err := cardano.NewAddress(receiverAddr)

	fmt.Println("receiver pub: ", hex.EncodeToString(receiver.Bytes()))
	if err != nil {
		log.Fatal("Address create:", err)
	}

	utxos, err := cli.UTxOs(sourceAddr)
	if err != nil {
		log.Fatal(err)
	}
	jsonUtxo, _ := json.Marshal(utxos)
	fmt.Println("utxos:", string(jsonUtxo))

	// Pick an Unspent Transaction Output as input for the transaction
	// For this example we'll pick the first utxo that has enough ADA for the
	// outputs and fee.

	builder := cardano.NewTxBuilder(
		pr,
	)

	// Send 5000000 lovelace or 5 ADA
	sendAmount := uint64((1000000))
	var firstMatchInput *cardano.TxInput
	// Loop through utxos to find first input with enough ADA
	for _, utxo := range utxos {
		if utxo.Amount.MultiAsset != nil && len(utxo.Amount.MultiAsset.Keys()) > 0 {
			continue
		}
		sendAmount = uint64(utxo.Amount.Sub(cardano.NewValue(cardano.Coin(200000))).Coin)
		minRequired := sendAmount + 200000
		if uint64(utxo.Amount.Coin) >= uint64(minRequired) {
			firstMatchInput = cardano.NewTxInput(utxo.TxHash, uint(utxo.Index), utxo.Amount)
			break
		}
	}

	builder.AddInputs(firstMatchInput)
	// Add a transaction output with the receiver's address and amount of 5 min
	out := cardano.NewTxOutput(
		receiver,
		cardano.NewValue(cardano.Coin(sendAmount)),
		//cardano.NewValueWithAssets(cardano.Coin(sendAmount), newAsset.MultiAsset()),
		)
	out.Amount.Coin = builder.MinCoinsForTxOut(out)
	builder.AddOutputs(out,
	)

	// Query tip from a node on the network. This is to get the current slot
	// and compute TTL of transaction.
	tip, err := cli.Tip()
	if err != nil {
		log.Fatal(err)
	}

	// Set TTL for 5 min into the future
	builder.SetTTL(uint64(tip.Slot) + uint64(300))
	fmt.Println("tip.Slot:", tip.Slot)
	//txRaw.Body.TTL = cardano.NewUint64(uint64(tip.Slot) + uint64(3000))

	// Route back the change to the source address
	// This is equivalent to adding an output with the source address and change amount
	builder.AddChangeIfNeeded(sourceAddr)

	// Build loops through the witness private keys and signs the transaction body hash
	builder.Sign(utxoPrvKey.PrvKey())
	txFinal, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	var newTx cardano.Tx
	(&newTx).UnmarshalCBOR(txFinal.Bytes())
	jdata, err := json.Marshal(newTx)
	fmt.Println(string(jdata), err)

	txHex := txFinal.Hex()

	fmt.Println(txHex, err)
	txid, err := txFinal.Hash()
	fmt.Println(hex.EncodeToString(txid), err)


	//txHash, err := cli.SubmitTx(txFinal)
	if err != nil {
		log.Fatal(err)
	}
	//, err := txFinal.Hash()
	//fmt.Println(txHash.String(), err)
}


