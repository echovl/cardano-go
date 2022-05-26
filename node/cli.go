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

// Cli implements Node using cardano-cli and a local node.
type Cli struct {
	network types.Network
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

// NewCli returns a new instance of Cli
func NewCli(network types.Network) Node {
	return &Cli{network: network}
}

func (c *Cli) runCommand(args ...string) ([]byte, error) {
	out := &bytes.Buffer{}

	if c.network == types.Mainnet {
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

func (c *Cli) UTXOs(address types.Address) ([]tx.Utxo, error) {
	out, err := c.runCommand("query", "utxo", "--address", string(address))
	if err != nil {
		return nil, err
	}

	utxos := []tx.Utxo{}
	lines := strings.Split(string(out), "\n")

	if len(lines) < 3 {
		return utxos, nil
	}

	for _, line := range lines[2 : len(lines)-1] {
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

func (c *Cli) Tip() (NodeTip, error) {
	out, err := c.runCommand("query", "tip")
	if err != nil {
		return NodeTip{}, err
	}

	cliTip := &cliTip{}
	if err = json.Unmarshal(out, cliTip); err != nil {
		return NodeTip{}, err
	}

	return NodeTip{
		Epoch: cliTip.Epoch,
		Block: cliTip.Block,
		Slot:  cliTip.Slot,
	}, nil
}

func (c *Cli) SubmitTx(tx tx.Transaction) (*types.Hash32, error) {
	txOut := cliTx{
		CborHex: tx.CborHex(),
	}

	txFile, err := ioutil.TempFile(os.TempDir(), "tx_")
	if err != nil {
		return nil, err
	}
	defer os.Remove(txFile.Name())

	if err := json.NewEncoder(txFile).Encode(txOut); err != nil {
		return nil, err
	}

	var txHash types.Hash32
	out, err := c.runCommand("transaction", "submit", "--tx-file", txFile.Name())
	if err != nil {
		return nil, err
	}
	copy(txHash[:], out)

	return &txHash, nil
}

func (c *Cli) Network() types.Network {
	return c.network
}
