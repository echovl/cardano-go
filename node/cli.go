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
	network    types.Network
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

func NewCli(network types.Network) Node {
	return &Cli{network: network}
}

func (cli *Cli) runCommand(args ...string) ([]byte, error) {
	out := &bytes.Buffer{}

	if cli.network == types.Mainnet {
		args = append(args, "--mainnet")
	} else {
		args = append(args, "--testnet-magic", strconv.Itoa(protocolMagic))
	}

	cmd := exec.Command("cardano-cli", args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

//TODO: add ability to use mainnet and testnet
func (cli *Cli) UTXOs(address types.Address) ([]tx.Utxo, error) {
	out, err := cli.runCommand("query", "utxo", "--address", string(address))
	if err != nil {
		return nil, err
	}

	utxos := []tx.Utxo{}
	lines := strings.Split(string(out), "\n")

	if len(lines) < 3 {
		return nil, fmt.Errorf("malformed cli response")
	}

	for _, line := range lines[2:] {
		args := strings.Fields(line)
		if len(args) < 4 {
			return nil, fmt.Errorf("malformed cli response")
		}
		txHash, err := types.NewHash32FromHex(args[0])
		if err != nil {
			return nil, err
		}
		index, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, err
		}
		amount, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, err
		}

		utxos = append(utxos, tx.Utxo{
			TxHash:  txHash,
			Index:   uint64(index),
			Amount:  types.Coin(amount),
			Address: address,
		})
	}

	return utxos, nil
}

//TODO: add ability to use mainnet and testnet
func (cli *Cli) Tip() (NodeTip, error) {
	out, err := cli.runCommand("query", "tip")
	if err != nil {
		return NodeTip{}, err
	}

	cliTip := &cliTip{}
	err = json.Unmarshal(out, cliTip)
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
		CborHex: tx.CborHex(),
	}

	txPayloadJson, err := json.Marshal(txPayload)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(txFileName, txPayloadJson, 777)
	if err != nil {
		return err
	}

	out, err := cli.runCommand("transaction", "submit", "--tx-file", txFileName)
	fmt.Print(out)

	err = os.Remove(txFileName)

	return err
}

func (cli *Cli) Network() types.Network {
	return cli.network
}
