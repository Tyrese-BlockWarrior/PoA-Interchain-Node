# Interchain Node [![Build Status](https://travis-ci.com/WeTrustPlatform/interchain-node.svg?token=zZKDmgBA4AupAdRbvfQN&branch=master)](https://travis-ci.com/WeTrustPlatform/interchain-node)

The interchain node allows ethereum users to transfer funds between a main chain and a side chain using multisig wallets.

## Dependencies

 * solc
 * abigen
 * go

### Dependencies of the development environment:

 * geth
 * bootnode
 * eth-contract-address

### GOPATH

In your .bashrc:

```
export GOPATH="$HOME/go"
export PATH="$GOPATH/bin:$PATH"
```

### Setup dependencies on OSX:

```
brew tap ethereum/ethereum
brew install ethereum
brew install solidity
go get github.com/kivutar/eth-contract-address
go install github.com/kivutar/eth-contract-address
```

`abigen` is supposed to be provided by the ethereum brew package on OSX. For some reasons, the installation of abigen doesn't always happen. (But it does install properly in Travis CI)
You may have to build it from source:

```
go get github.com/ethereum/go-ethereum
cd $GOPATH/src/github.com/ethereum/go-ethereum/cmd/abigen/
go build
sudo cp abigen /usr/local/bin/abigen
```

## Getting the source code

    go get github.com/WeTrustPlatform/poa-interchain-node

But as the repo is still private, you will have to clone the repo directly in your go path:

    mkdir -p $GOPATH/src/github.com/WeTrustPlatform/
    cd $GOPATH/src/github.com/WeTrustPlatform/
    git clone --recurse-submodules git@github.com:WeTrustPlatform/poa-interchain-node.git
    cd interchain-node
    go generate ./...
    go get -t -v ./...

## Run the test suite

    go test -cover -v ./...

## Run the development environment

The interchain node comes with a helper in the form of a Makefile. This Makefile is not required to build the interchain node and use it. It is there to test the interchain node with the real geth. It will setup 3 instances of geth, 2 chains, and a few test accounts.

If you don't need to test the interchain node on geth, please skip all of the following steps and rely on the test suite only.

In 3 separate terminals:

```
make run_mainchain
make run_sidechain
make run_sidechain2
```

Wait 10 seconds between each call

## Deploy the multisig wallet on both chains

    make mainchain_wallet sidechain_wallet

## Sending ether to arbitrary addresses offchain

Our multisig wallets have a function called `deposit` which allows you to transfer ether from your address on the chain A to a receiver on the chain B.

Instead of paying to the wallet address, you call the `deposit` method using the geth console or our utility program `icn-deposit.go`:

    go run cmd/icn-deposit/main.go --mainchain --keyjson=mainchain/keystore/<your_key_json> --password="dummy" --endpoint=mainchain/geth.ipc --wallet=`cat mainchain/wallet` --receiver=`cat sidechain/sealer` --value="25000000000000000000"

Usage:

```
Usage:
  main [OPTIONS]

Application Options:
  -m, --mainchain  Target the main chain wallet
  -s, --sidechain  Target the side chain wallet
  -k, --keyjson=   Path to the JSON private key file of the sealer
  -p, --password=  Passphrase needed to unlock the sealer's JSON key
  -e, --endpoint=  URL or path of the origin chain endpoint
  -w, --wallet=    Ethereum address of the multisig wallet on the origin chain
  -r, --receiver=  Ethereum address of the receiver on the target chain
  -v, --value=     Value (wei) to transfer to the receiver

Help Options:
  -h, --help       Show this help message
```

The interchain node will notice your call and mirror the transaction on the other chain.

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
