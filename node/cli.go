package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

type Cli struct {
	socketPath string
}

type cliTip struct {
	Epoch uint64
	Hash  string
	Slot  uint64
	Block uint64
	Era   string
}

type cliTx struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CborHex     string `json:"cborHex"`
}

func NewCli() Node {
	return &Cli{}
}

//TODO: add ability to use mainnet and testnet
func (cli *Cli) QueryUtxos(address types.Address) ([]tx.Utxo, error) {
	out, err := runCommand("cardano-cli", "query", "utxo", "--address", string(address), "--testnet-magic", "1097911063")
	if err != nil {
		return nil, err
	}

	counter := 0
	utxos := []tx.Utxo{}
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

			txId := tx.TransactionID(args[0])
			index, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return nil, err
			}
			amount, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return nil, err
			}

			utxos = append(utxos, tx.Utxo{
				TxId:    txId,
				Index:   index,
				Amount:  types.Coin(amount),
				Address: address,
			})
		}
		counter++
	}

	return utxos, nil
}

//TODO: add ability to use mainnet and testnet
func (cli *Cli) QueryTip() (NodeTip, error) {
	out, err := runCommand("cardano-cli", "query", "tip", "--testnet-magic", "1097911063")
	if err != nil {
		return NodeTip{}, err
	}

	cliTip := &cliTip{}
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
func (cli *Cli) SubmitTx(tx tx.Transaction) error {
	const txFileName = "txsigned.temp"
	txPayload := cliTx{
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
