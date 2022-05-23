package cardano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	testnetMagic = 1097911063
)

type cardanoNode interface {
	QueryUtxos(Address) ([]Utxo, error)
	QueryTip() (NodeTip, error)
	SubmitTx(Transaction) error
}

type Utxo struct {
	Address Address
	TxId    TransactionID
	Amount  uint64
	Index   uint64
}

type NodeTip struct {
	Epoch uint64
	Block uint64
	Slot  uint64
}

type cardanoCli struct {
	socketPath string
}

type cardanoCliTip struct {
	Epoch uint64
	Hash  string
	Slot  uint64
	Block uint64
	Era   string
}

type cardanoCliTx struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CborHex     string `json:"cborHex"`
}

func newCli() *cardanoCli {
	return &cardanoCli{}
}

//TODO: add ability to use mainnet and testnet
func (cli *cardanoCli) QueryUtxos(address Address) ([]Utxo, error) {
	out, err := runCommand("cardano-cli", "query", "utxo", "--address", string(address), "--testnet-magic", "1097911063")
	if err != nil {
		return nil, err
	}

	counter := 0
	utxos := []Utxo{}
	for {
		line, err := out.ReadString(byte('\n'))
		if err != nil {
			break
		}
		if counter >= 2 {
			args := strings.Fields(line)
			if len(args) < 4 {
				return nil, fmt.Errorf("malformed cli response")
			}

			txId := TransactionID(args[0])
			index, err := ParseUint64(args[1])
			if err != nil {
				return nil, err
			}
			amount, err := ParseUint64(args[2])
			if err != nil {
				return nil, err
			}

			utxos = append(utxos, Utxo{
				TxId:    txId,
				Index:   index,
				Amount:  amount,
				Address: address,
			})
		}
		counter++
	}

	return utxos, nil
}

//TODO: add ability to use mainnet and testnet
func (cli *cardanoCli) QueryTip() (NodeTip, error) {
	out, err := runCommand("cardano-cli", "query", "tip", "--testnet-magic", "1097911063")
	if err != nil {
		return NodeTip{}, err
	}

	cliTip := &cardanoCliTip{}
	err = json.Unmarshal(out.Bytes(), cliTip)
	if err != nil {
		return NodeTip{}, err
	}

	return NodeTip{
		Epoch: cliTip.Epoch,
		Block: cliTip.Block,
		Slot:  cliTip.Slot,
	}, nil
}

//TODO: add ability to use mainnet and testnet
func (cli *cardanoCli) SubmitTx(tx Transaction) error {
	const txFileName = "txsigned.temp"
	txPayload := cardanoCliTx{
		Type:        "Tx MaryEra",
		Description: "",
		CborHex:     tx.CborHex(),
	}

	txPayloadJson, err := json.Marshal(txPayload)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(txFileName, txPayloadJson, 777)
	if err != nil {
		return err
	}

	out, err := runCommand("cardano-cli", "transaction", "submit", "--tx-file", txFileName, "--testnet-magic", "1097911063")
	fmt.Print(out.String())

	err = os.Remove(txFileName)

	return err
}

func runCommand(cmd string, arg ...string) (*bytes.Buffer, error) {
	out := &bytes.Buffer{}
	command := exec.Command(cmd, arg...)
	command.Stdout = out
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return nil, err
	}

	return out, nil
}
