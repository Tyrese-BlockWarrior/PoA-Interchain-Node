# Interchain Node

The interchain node allows ethereum users to transfer funds between a main chain and a side chain using multisig wallets.

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

## Sending ether to arbitrary addresses offchain

The multisig wallet has a function called `deposit` which allows you to transfer ether from your address on the chain A to a receiver on the chain B.

Instead of paying to the wallet address, you call the `deposit` method using geth console or our utility program `icn-deposit.go`:

    go run cmd/icn-deposit/main.go -mainchain -keyjson=mainchain/keystore/<your_key_json> -password=dummy -endpoint=mainchain/geth.ipc -wallet=`cat mainchain/miner` -receiver=`cat sidechain/sealer` -value="25000000000000000000"

Usage:

```
  -endpoint string
    	URL or path of the origin chain endpoint
  -keyjson string
    	Path to the JSON private key file of the user
  -mainchain
    	Deploy the main chain wallet
  -password string
    	Passphrase needed to unlock the user's JSON key
  -receiver string
    	Ethereum address of the receiver on the target chain
  -sidechain
    	Deploy the side chain wallet
  -value string
    	Value (wei) to transfer to the receiver
  -wallet string
    	Ethereum address of the multisig wallet on the origin chain
```

The interchain node will notice your call and mirror the transaction on the other chain.