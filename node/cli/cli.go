package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/echovl/cardano-go/node"
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

// CardanoCli implements Node using cardano-cli and a local node.
type CardanoCli struct {
	network types.Network
}

type tip struct {
	Epoch int
	Hash  string
	Slot  int
	Block int
	Era   string
}

type cliTx struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CborHex     string `json:"cborHex"`
}

// NewNode returns a new instance of CardanoCli.
func NewNode(network types.Network) node.Node {
	return &CardanoCli{network: network}
}

func (c *CardanoCli) runCommand(args ...string) ([]byte, error) {
	out := &bytes.Buffer{}

	if c.network == types.Mainnet {
		args = append(args, "--mainnet")
	} else {
		args = append(args, "--testnet-magic", strconv.Itoa(node.ProtocolMagic))
	}

	cmd := exec.Command("cardano-cli", args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (c *CardanoCli) UTXOs(addr types.Address) ([]tx.UTXO, error) {
	out, err := c.runCommand("query", "utxo", "--address", addr.String())
	if err != nil {
		return nil, err
	}

	utxos := []tx.UTXO{}
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

		utxos = append(utxos, tx.UTXO{
			Spender: addr,
			TxHash:  txHash,
			Index:   uint64(index),
			Amount:  types.Coin(amount),
		})
	}

	return utxos, nil
}

func (c *CardanoCli) Tip() (*node.NodeTip, error) {
	out, err := c.runCommand("query", "tip")
	if err != nil {
		return nil, err
	}

	cliTip := &tip{}
	if err = json.Unmarshal(out, cliTip); err != nil {
		return nil, err
	}

	return &node.NodeTip{
		Epoch: cliTip.Epoch,
		Block: cliTip.Block,
		Slot:  cliTip.Slot,
	}, nil
}

func (c *CardanoCli) SubmitTx(tx *tx.Transaction) (*types.Hash32, error) {
	txOut := cliTx{
		Type:    "Witnessed Tx AlonzoEra",
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

	out, err := c.runCommand("transaction", "submit", "--tx-file", txFile.Name())
	if err != nil {
		return nil, errors.New(string(out))
	}

	txHash, err := tx.Hash()
	if err != nil {
		return nil, err
	}

	return &txHash, nil
}

func (c *CardanoCli) Network() types.Network {
	return c.network
}
