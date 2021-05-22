package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/echovl/cardano-wallet/wallet"
)

type NodeCli struct{}

type CliTip struct {
	Epoch uint64
	Hash  string
	Slot  uint64
	Block uint64
	Era   string
}

func (cli *NodeCli) QueryUtxos(address wallet.Address) ([]wallet.Utxo, error) {
	out, err := runCommand("cardano-cli", "query", "utxo", "--address", string(address), "--testnet-magic", "1097911063")
	if err != nil {
		return nil, err
	}

	counter := 0
	utxos := []wallet.Utxo{}
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

			txId := wallet.TxId(args[0])
			index, err := parseUint64(args[1])
			if err != nil {
				return nil, err
			}
			amount, err := parseUint64(args[2])
			if err != nil {
				return nil, err
			}

			utxos = append(utxos, wallet.Utxo{
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

func (cli *NodeCli) QueryTip() (wallet.NodeTip, error) {
	out, err := runCommand("cardano-cli", "query", "tip", "--testnet-magic", "1097911063")
	if err != nil {
		return wallet.NodeTip{}, err
	}

	cliTip := &CliTip{}
	err = json.Unmarshal(out.Bytes(), cliTip)
	if err != nil {
		return wallet.NodeTip{}, err
	}

	return wallet.NodeTip{
		Epoch: cliTip.Epoch,
		Block: cliTip.Block,
		Slot:  cliTip.Slot,
	}, nil
}

func runCommand(cmd string, arg ...string) (*bytes.Buffer, error) {
	out := &bytes.Buffer{}
	command := exec.Command(cmd, arg...)
	command.Stdout = out

	err := command.Run()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func parseUint64(s string) (uint64, error) {
	parsed, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
