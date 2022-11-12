package cardanocli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/echovl/cardano-go"
)

var socketPath, fallbackSocketPath string

func AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&socketPath, "cardano-node-socket-path", "", "")
	fs.StringVar(&fallbackSocketPath, "fallback-cardano-node-socket-path", "", "")
}

func availableAsSocket(fn string) bool {
	fi, err := os.Stat(fn)
	if err != nil {
		return false
	}
	if (fi.Mode().Type() & os.ModeSocket) != os.ModeSocket {
		return false
	}
	return true
}

func getSocketPathToUse() string {
	sp := socketPath
	if sp != "" && availableAsSocket(sp) {
		return sp
	}
	sp = fallbackSocketPath
	if sp != "" && availableAsSocket(sp) {
		return sp
	}
	return os.Getenv("CARDANO_NODE_SOCKET_PATH")
}

// CardanoCli implements Node using cardano-cli and a local node.
type CardanoCli struct {
	ctx     context.Context
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
func NewNode(network cardano.Network, rest ...any) cardano.Node {
	if len(rest) > 0 {
		return &CardanoCli{network: network, ctx: rest[0].(context.Context)}
	}
	return &CardanoCli{network: network}
}

func (c *CardanoCli) runCommand(args ...string) (string, error) {
	out := &strings.Builder{}

	if c.network == cardano.Mainnet {
		args = append(args, "--mainnet")
	} else {
		args = append(args, "--testnet-magic", strconv.Itoa(cardano.ProtocolMagic))
	}

	var cmd *exec.Cmd
	if c.ctx == nil {
		cmd = exec.Command("cardano-cli", args...)
	} else {
		cmd = exec.CommandContext(c.ctx, "cardano-cli", args...)
	}
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}

func (c *CardanoCli) DoCommand(args ...string) (string, error) {
	return c.runCommand(args...)
}

func (c *CardanoCli) UTxOs(addr cardano.Address) ([]cardano.UTxO, error) {
	out, err := c.runCommand("query", "utxo", "--address", addr.Bech32())
	if err != nil {
		return nil, err
	}

	utxos := []cardano.UTxO{}
	lines := strings.Split(out, "\n")

	if len(lines) < 3 {
		return utxos, nil
	}

	for _, line := range lines[2 : len(lines)-1] {
		amount := cardano.NewValue(0)
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
		lovelace, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, err
		}
		amount.Coin = cardano.Coin(lovelace)

		assets := strings.Split(line, "+")
		for _, asset := range assets[1 : len(assets)-1] {
			args := strings.Fields(asset)
			quantity := args[0]
			unit := strings.ReplaceAll(args[1], ".", "")
			unitBytes, err := hex.DecodeString(unit)
			if err != nil {
				return nil, err
			}
			policyID := cardano.NewPolicyIDFromHash(unitBytes[:28])
			assetName := string(unitBytes[28:])
			assetValue, err := strconv.ParseUint(quantity, 10, 64)
			if err != nil {
				return nil, err
			}
			currentAssets := amount.MultiAsset.Get(policyID)
			if currentAssets != nil {
				currentAssets.Set(
					cardano.NewAssetName(assetName),
					cardano.BigNum(assetValue),
				)
			} else {
				amount.MultiAsset.Set(
					policyID,
					cardano.NewAssets().
						Set(
							cardano.NewAssetName(string(assetName)),
							cardano.BigNum(assetValue),
						),
				)
			}
		}

		utxos = append(utxos, cardano.UTxO{
			Spender: addr,
			TxHash:  txHash,
			Index:   uint64(index),
			Amount:  amount,
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
	if err = json.Unmarshal([]byte(out), cliTip); err != nil {
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
		return nil, errors.New(out)
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
		return nil, errors.New(out)
	}

	var cparams protocolParameters
	if err := json.Unmarshal([]byte(out), &cparams); err != nil {
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
