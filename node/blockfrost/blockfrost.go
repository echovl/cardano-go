package blockfrost

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go/node"
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

// BlockfrostNode implements Node using the blockfrost API.
type BlockfrostNode struct {
	client    blockfrost.APIClient
	projectID string
	network   types.Network
}

// NewNode returns a new instance of BlockfrostNode.
func NewNode(network types.Network, projectID string) node.Node {
	return &BlockfrostNode{
		network:   network,
		projectID: projectID,
		client: blockfrost.NewAPIClient(blockfrost.APIClientOptions{
			ProjectID: projectID,
		}),
	}
}

func (b *BlockfrostNode) UTXOs(addr types.Address) ([]tx.UTXO, error) {
	butxos, err := b.client.AddressUTXOs(context.Background(), addr.String(), blockfrost.APIQueryParams{})
	if err != nil {
		return nil, err
	}

	utxos := make([]tx.UTXO, len(butxos))

	for i, butxo := range butxos {
		txHash, err := types.NewHash32(butxo.TxHash)
		if err != nil {
			return nil, err
		}

		var amount uint64
		for _, a := range butxo.Amount {
			if a.Unit == "lovelace" {
				lovelace, err := strconv.ParseUint(a.Quantity, 10, 64)
				if err != nil {
					return nil, err
				}
				amount += lovelace
			}
		}

		utxos[i] = tx.UTXO{
			Spender: addr,
			TxHash:  txHash,
			Amount:  types.Coin(amount),
			Index:   uint64(butxo.OutputIndex),
		}
	}

	return utxos, nil
}

func (b *BlockfrostNode) Tip() (*node.NodeTip, error) {
	block, err := b.client.BlockLatest(context.Background())
	if err != nil {
		return nil, err
	}

	return &node.NodeTip{
		Block: block.Height,
		Epoch: block.Epoch,
		Slot:  block.Slot,
	}, nil
}

func (b *BlockfrostNode) SubmitTx(tx *tx.Transaction) (*types.Hash32, error) {
	url := fmt.Sprintf("https://cardano-%s.blockfrost.io/api/v0/tx/submit", b.network.String())
	txBytes := tx.Bytes()

	req, err := http.NewRequest("POST", url, bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Add("project_id", b.projectID)
	req.Header.Add("Content-Type", "application/cbor")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	txHash, err := tx.Hash()
	if err != nil {
		return nil, err
	}

	return &txHash, nil
}

func (b *BlockfrostNode) Network() types.Network {
	return b.network
}
