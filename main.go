/*
Copyright (C) 2018 WeTrustPlatform

This file is part of interchain-node.

interchain-node is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

interchain-node is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with interchain-node.  If not, see <http://www.gnu.org/licenses/>.
*/

package icn

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/WeTrustPlatform/interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/miguelmota/go-solidity-sha3"
)

//go:generate mkdir -p bind/mainchain/
//go:generate mkdir -p bind/sidechain/
//go:generate abigen --sol interchain-node-contracts/contracts/MainChain.sol --pkg mainchain --out bind/mainchain/main.go
//go:generate abigen --sol interchain-node-contracts/contracts/SideChain.sol --pkg sidechain --out bind/sidechain/main.go

// MsgHash returns the sha3 sum of 0x19, contractAddress, txHash, toAddress, value, data and version
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
		solsha3.Bytes32(string(txHash[:])),
		solsha3.Address(toAddress),
		solsha3.Int256(value),
		solsha3.String(data),
	)

	msgHash.SetBytes(hash)

	return msgHash
}

// DepositInfo groups a DepositEvent and a Transaction to be processed
type DepositInfo struct {
	Event  DepositEvent
	TxHash common.Hash
}

// DepositEvent is used in unpacking events
type DepositEvent struct {
	Sender   common.Address
	Receiver common.Address
	Value    *big.Int
}

// ParseSignature parses a ECDSA signature and returns v, r, s
func ParseSignature(sig []byte) (v uint8, r, s common.Hash) {
	r = common.BytesToHash(sig[0:32])
	s = common.BytesToHash(sig[32:64])
	v = uint8(sig[64:65][0] + 27)

	return
}

// SubmitSignatureMC submits a signature on the sidechain that will be later checked on the mainchain
func SubmitSignatureMC(
	ctx context.Context,
	sideChainWalletAddress common.Address,
	auth *bind.TransactOpts,
	sc *sidechain.SideChain,
	event *sidechain.SideChainDeposit,
	key *ecdsa.PrivateKey,
) (*types.Transaction, error) {
	// Create the message hash
	var data []byte

	msgHash := MsgHash(sideChainWalletAddress, event.Raw.TxHash, event.To, event.Value, data, 1)

	// Sign the message hash
	sig, err := crypto.Sign(msgHash.Bytes(), key)
	if err != nil {
		return nil, errors.New("Sign failed: " + err.Error())
	}

	// Parse the signature
	v, r, s := ParseSignature(sig)

	// Submit the signature
	return sc.SubmitSignatureMC(auth, event.Raw.TxHash, event.To, event.Value, data, v, r, s)
}
