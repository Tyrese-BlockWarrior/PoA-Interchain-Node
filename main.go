package icn

import (
	"context"
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/miguelmota/go-solidity-sha3"
)

//go:generate mkdir -p bind/mainchain/
//go:generate mkdir -p bind/sidechain/
//go:generate abigen --sol interchain-node-contracts/contracts/MainChain.sol --pkg mainchain --out bind/mainchain/main.go
//go:generate abigen --sol interchain-node-contracts/contracts/SideChain.sol --pkg sidechain --out bind/sidechain/main.go

// MsgHash returns the sha3 sum of contractAddress, txHash, toAddress, value, data and version
func MsgHash(
	contractAddress common.Address,
	txHash common.Hash,
	toAddress common.Address,
	value *big.Int,
	data []byte,
	version uint8) common.Hash {
	var msgHash common.Hash

	hash := solsha3.SoliditySHA3(
		[]byte{0x19},
		solsha3.Uint8(version),
		solsha3.Address(contractAddress),
		solsha3.Bytes32(txHash.Str()),
		solsha3.Address(toAddress),
		solsha3.Int256(value),
		solsha3.String(data),
	)

	msgHash.SetBytes(hash)

	return msgHash
}

// DepositEvent is used in unpacking events
type DepositEvent struct {
	Sender   common.Address
	Receiver common.Address
	Value    *big.Int
}

// GetDepositEvent loops over the logs and returns the depositEvent
func GetDepositEvent(abi abi.ABI, logs []*types.Log) DepositEvent {

	// There should be only one deposit event in the logs
	for _, l := range logs {
		var depositEvent DepositEvent

		err := abi.Unpack(&depositEvent, "Deposit", l.Data)
		if err != nil {
			log.Printf("Event log unpack error: %v", err)
			continue
		}

		// A valid deposit event has a sender and a receiver in Topics
		if len(l.Topics) != 3 {
			return DepositEvent{}
		}

		// Indexed attributes go in l.Topics instead of l.Data
		depositEvent.Sender = common.BytesToAddress(l.Topics[1].Bytes())
		depositEvent.Receiver = common.BytesToAddress(l.Topics[2].Bytes())
		return depositEvent
	}

	return DepositEvent{}
}

// ParseSignature parses a ECDSA signature and returns v, r, s
func ParseSignature(sig []byte) (v uint8, r, s common.Hash) {
	r = common.BytesToHash(sig[0:32])
	s = common.BytesToHash(sig[32:64])
	v = uint8(sig[64:65][0] + 27)

	return
}

// GetLogs returns the logs for a transaction
func GetLogs(ctx context.Context, client *ethclient.Client, tx *types.Transaction) ([]*types.Log, error) {
	// Get the transaction receipt
	receipt, err := client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, err
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return nil, errors.New("Receipt status is ReceiptStatusFailed")
	}

	return receipt.Logs, nil
}
