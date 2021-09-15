package cardano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	testnetMagic = 1097911063
)

type cardanoNode interface {
	QueryUtxos(Address) ([]Utxo, error)
	QueryTip() (NodeTip, error)
	SubmitTx(transaction) error
	ProtocolParameters() (Protocol, error)
}

type Utxo struct {
	Address Address
	TxId    transactionID
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

type Protocol struct {
	TxFeePerByte           uint64
	MinUTxOValue           uint64
	StakePoolDeposit       uint64
	UtxoCostPerWord        uint64
	Decentralization       int
	PoolRetireMaxEpoch     int
	CollateralPercentage   int
	StakePoolTargetNum     int
	MaxBlockBodySize       uint64
	MaxTxSize              uint64
	TreasuryCut            float64
	MinPoolCost            uint64
	MaxCollateralInputs    int
	MaxValueSize           int
	MaxBlockExecutionUnits struct {
		Memory uint64
		Steps  uint64
	}
	MaxBlockHeaderSize  int
	MaxTxExecutionUnits struct {
		Memory uint64
		Steps  uint64
	}
	ProtocolVersion struct {
		Minor int
		Major int
	}
	TxFeeFixed          uint64
	StakeAddressDeposit uint64
	MonetaryExpansion   float64
	PoolPledgeInfluence float64
	ExecutionUnitPrices struct {
		PriceSteps  float64
		PriceMemory float64
	}
}

func newCli(socketPath string) *cardanoCli {
	return &cardanoCli{socketPath: socketPath}
}

//TODO: add ability to use mainnet and testnet
func (cli *cardanoCli) QueryUtxos(address Address) ([]Utxo, error) {
	out, err := cli.runCommand("cardano-cli", "query", "utxo", "--address", string(address), "--testnet-magic", strconv.Itoa(testnetMagic))
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

			txId := transactionID(args[0])
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
	out, err := cli.runCommand("cardano-cli", "query", "tip", "--testnet-magic", strconv.Itoa(testnetMagic))
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
func (cli *cardanoCli) SubmitTx(tx transaction) error {
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

	out, err := cli.runCommand("cardano-cli", "transaction", "submit", "--tx-file", txFileName, "--testnet-magic", strconv.Itoa(testnetMagic))
	if err != nil {
		return err
	}
	fmt.Print(out.String())

	err = os.Remove(txFileName)

	return err
}

func (cli *cardanoCli) ProtocolParameters() (Protocol, error) {
	var data Protocol
	out, err := cli.runCommand("cardano-cli", "query", "protocol-parameters", "--testnet-magic", strconv.Itoa(testnetMagic))
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(out.Bytes(), &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func (cli *cardanoCli) runCommand(cmd string, arg ...string) (*bytes.Buffer, error) {
	out := &bytes.Buffer{}
	command := exec.Command(cmd, arg...)
	command.Stdout = out
	command.Stderr = os.Stderr
	command.Env = os.Environ()
	command.Env = append(command.Env, "CARDANO_NODE_SOCKET_PATH="+cli.socketPath)
	err := command.Run()
	if err != nil {
		return nil, err
	}

	return out, nil
}
