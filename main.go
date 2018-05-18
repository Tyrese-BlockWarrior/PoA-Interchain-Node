package icn

import (
	"context"
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/miguelmota/go-solidity-sha3"
)

//go:generate mkdir -p ../../bind/mainchain/
//go:generate mkdir -p ../../bind/sidechain/
//go:generate abigen --sol ../../interchain-node-contracts/contracts/MainChain.sol --pkg mainchain --out ../../bind/mainchain/main.go
//go:generate abigen --sol ../../interchain-node-contracts/contracts/SideChain.sol --pkg sidechain --out ../../bind/sidechain/main.go

// MsgHash returns the sha3 sum of txHash, destination, value, data and version
func MsgHash(txHash common.Hash, destination common.Address, value *big.Int, data []byte, version uint8) common.Hash {
	var msgHash common.Hash

	hash := solsha3.SoliditySHA3(
		solsha3.Bytes32(txHash.Str()),
		solsha3.Address(destination.String()),
		solsha3.Int256(value),
		solsha3.String(string(data[:])),
		solsha3.Uint8(version),
	)

	msgHash.SetBytes(hash)

	return msgHash
}

// ToByte32 convers a variable lenght byte slice to a fixed lenght byte slice
func ToByte32(in []byte) (out [32]byte) {
	copy(out[:], in)
	return out
}

// DepositEvent is used in unpacking events
type DepositEvent struct {
	Sender   common.Address
	Receiver common.Address
	Value    *big.Int
}

// GetDepositEvent loops over the logs and returns the depositEvent
func GetDepositEvent(abi abi.ABI, logs []*types.Log) DepositEvent {

	var depositEvent DepositEvent

	// There should be only one deposit event in the logs
	for _, l := range logs {
		err := abi.Unpack(&depositEvent, "Deposit", l.Data)
		if err != nil {
			log.Printf("Event log unpack error: %v", err)
			continue
		}

		// Indexed attributes go in l.Topics instead of l.Data
		depositEvent.Sender = common.BytesToAddress(l.Topics[1].Bytes())
		depositEvent.Receiver = common.BytesToAddress(l.Topics[2].Bytes())
	}

	return depositEvent
}

// GetRawSignature creates the transaction, signs it, and returns the v, r, s
func GetRawSignature(
	ctx context.Context,
	auth *bind.TransactOpts,
	value *big.Int,
	key *keystore.Key,
	destination common.Address,
	mainchainClient *ethclient.Client,
	data []byte) (uint8, [32]byte, [32]byte, error) {

	nonce, err := mainchainClient.NonceAt(ctx, auth.From, nil)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, errors.New("NonceAt failed: " + err.Error())
	}

	var gasLimit uint64 = 100000
	gasPrice := big.NewInt(20000000000) // 20 gwei
	newTx := types.NewTransaction(nonce, destination, value, gasLimit, gasPrice, data)

	signer := types.HomesteadSigner{}

	signedTx, err := types.SignTx(newTx, signer, key.PrivateKey)
	if err != nil {
		return 0, [32]byte{}, [32]byte{},

			errors.New("SignTx failed: " + err.Error())
	}

	v, r, s := signedTx.RawSignatureValues()

	return uint8(v.Uint64()), ToByte32(r.Bytes()), ToByte32(s.Bytes()), nil
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
