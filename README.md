# Interchain Node [![Build Status](https://travis-ci.com/WeTrustPlatform/interchain-node.svg?token=zZKDmgBA4AupAdRbvfQN&branch=master)](https://travis-ci.com/WeTrustPlatform/interchain-node)

The interchain node allows ethereum users to transfer funds between a main chain and a side chain using multisig wallets.

## Dependencies

 * solc
 * abigen
 * go

### Dependencies of the development environment:

 * geth
 * bootnode

### Setup dependencies on OSX:

```
brew tap ethereum/ethereum
brew install ethereum
brew install solidity
```

## Getting the source code

    go get github.com/WeTrustPlatform/interchain-node

But as the repo is still private, you will have to clone the repo directly in your go path:

    cd $GOPATH
    mkdir -p src/github.com/WeTrustPlatform/
    cd src/github.com/WeTrustPlatform/
    git clone git@github.com:WeTrustPlatform/interchain-node.git
    cd interchain-node

## Run the test suite

    go test -cover -v ./...

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

    go run cmd/icn/main.go -k sidechain/keystore/<sealer_key_json> --mainchainendpoint=mainchain/geth.ipc --sidechainendpoint=sidechain/geth.ipc -p dummy --mainchainwallet=`cat mainchain/wallet` --sidechainwallet=`cat sidechain/wallet`

```
Usage:
  main [OPTIONS]

Application Options:
  -m, --mainchain          Watch the main chain
  -s, --sidechain          Watch the side chain
  -k, --keyjson=           Path to the JSON private key file of the sealer
  -p, --password=          Passphrase needed to unlock the sealer's JSON key
      --mainchainendpoint= URL or path of the main chain endpoint
      --sidechainendpoint= URL or path of the side chain endpoint
      --mainchainwallet=   Ethereum address of the multisig wallet on the main chain
      --sidechainwallet=   Ethereum address of the multisig wallet on the side chain

Help Options:
  -h, --help               Show this help message
```

## Sending ether to arbitrary addresses offchain

Our multisig wallets have a function called `deposit` which allows you to transfer ether from your address on the chain A to a receiver on the chain B.

Instead of paying to the wallet address, you call the `deposit` method using the geth console or our utility program `icn-deposit.go`:

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