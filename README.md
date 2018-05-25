# Interchain Node [![Build Status](https://travis-ci.com/WeTrustPlatform/interchain-node.svg?token=zZKDmgBA4AupAdRbvfQN&branch=master)](https://travis-ci.com/WeTrustPlatform/interchain-node)

The interchain node allows ethereum users to transfer funds between a main chain and a side chain using multisig wallets.

It is composed of two programs:

 * icn-mc2sc: Watches the main chain and react on the side chain
 * icn-sc2mc: Watches the side chain and react on the main chain (Unimplemented yet)

## Dependencies

 * solc
 * abigen
 * go
 * geth

### Dependencies for developers:

 * bootnode
 * eth-contract-address

### Setup dependencies on OSX:

```
brew install geth
brew install solidity
go install github.com/kivutar/eth-contract-address
```

## Getting the source code

    go get github.com/WeTrustPlatform/interchain-node

## Run the development environment

In 3 separate terminals:

```
make run_mainchain
make run_sidechain
make run_sidechain2
```

Wait 10 seconds between each call

## Deploy the multisig wallet on both chains

    make mainchain_wallet sidechain_wallet

## Run the interchain node

For each sealer, run the interchain node:

    go run cmd/icn/main.go -keyjson=sidechain/keystore/<sealer_key_json> -mainchainendpoint=mainchain/geth.ipc -sidechainendpoint=sidechain/geth.ipc -password=dummy -mainchainwallet=`cat mainchain/wallet` -sidechainwallet=`cat sidechain/wallet`

## Sending ether to arbitrary addresses offchain

Our multisig wallets have a function called `deposit` which allows you to transfer ether from your address on the chain A to a receiver on the chain B.

Instead of paying to the wallet address, you call the `deposit` method using geth console or our utility program `icn-deposit.go`:

    go run cmd/icn-deposit/main.go -mainchain -keyjson=mainchain/keystore/<your_key_json> -password=dummy -endpoint=mainchain/geth.ipc -wallet=`cat mainchain/wallet` -receiver=`cat sidechain/sealer` -value="25000000000000000000"

Usage:

```
  -endpoint string
    	URL or path of the origin chain endpoint
  -keyjson string
    	Path to the JSON private key file of the user
  -mainchain
    	Target the main chain wallet
  -password string
    	Passphrase needed to unlock the user's JSON key
  -receiver string
    	Ethereum address of the receiver on the target chain
  -sidechain
    	Target the side chain wallet
  -value string
    	Value (wei) to transfer to the receiver
  -wallet string
    	Ethereum address of the multisig wallet on the origin chain
```

The interchain node will notice your call and mirror the transaction on the other chain.