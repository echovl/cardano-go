package cardanocli

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

	"github.com/echovl/cardano-go"
)

// CardanoCli implements Node using cardano-cli and a local node.
type CardanoCli struct {
	network cardano.Network
}

type tip struct {
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

// NewNode returns a new instance of CardanoCli.
func NewNode(network cardano.Network) cardano.Node {
	return &CardanoCli{network: network}
}

func (c *CardanoCli) runCommand(args ...string) ([]byte, error) {
	out := &bytes.Buffer{}

	if c.network == cardano.Mainnet {
		args = append(args, "--mainnet")
	} else {
		args = append(args, "--testnet-magic", strconv.Itoa(cardano.ProtocolMagic))
	}

	cmd := exec.Command("cardano-cli", args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (c *CardanoCli) UTxOs(addr cardano.Address) ([]cardano.UTxO, error) {
	out, err := c.runCommand("query", "utxo", "--address", addr.Bech32())
	if err != nil {
		return nil, err
	}

	utxos := []cardano.UTxO{}
	lines := strings.Split(string(out), "\n")

	if len(lines) < 3 {
		return utxos, nil
	}

	for _, line := range lines[2 : len(lines)-1] {
		args := strings.Fields(line)
		if len(args) < 4 {
			return nil, fmt.Errorf("malformed cli response")
		}
		txHash, err := cardano.NewHash32(args[0])
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

		utxos = append(utxos, cardano.UTxO{
			Spender: addr,
			TxHash:  txHash,
			Index:   uint64(index),
			Amount:  cardano.Coin(amount),
		})
	}

	return utxos, nil
}

func (c *CardanoCli) Tip() (*cardano.NodeTip, error) {
	out, err := c.runCommand("query", "tip")
	if err != nil {
		return nil, err
	}

	cliTip := &tip{}
	if err = json.Unmarshal(out, cliTip); err != nil {
		return nil, err
	}

	return &cardano.NodeTip{
		Epoch: cliTip.Epoch,
		Block: cliTip.Block,
		Slot:  cliTip.Slot,
	}, nil
}

func (c *CardanoCli) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	txOut := cliTx{
		Type:    "Witnessed Tx AlonzoEra",
		CborHex: tx.Hex(),
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

type protocolParameters struct {
	MinFeeA          cardano.Coin `json:"txFeePerByte"`
	MinFeeB          cardano.Coin `json:"txFeeFixed"`
	CoinsPerUTXOWord cardano.Coin `json:"utxoCostPerWord"`
}

func (c *CardanoCli) ProtocolParams() (*cardano.ProtocolParams, error) {
	out, err := c.runCommand("query", "protocol-parameters")
	if err != nil {
		return nil, errors.New(string(out))
	}

	var cparams protocolParameters
	if err := json.Unmarshal(out, &cparams); err != nil {
		return nil, err
	}

	pparams := &cardano.ProtocolParams{
		MinFeeA:          cparams.MinFeeA,
		MinFeeB:          cparams.MinFeeB,
		CoinsPerUTXOWord: cparams.CoinsPerUTXOWord,
	}

	return pparams, nil
}

func (c *CardanoCli) Network() cardano.Network {
	return c.network
}
