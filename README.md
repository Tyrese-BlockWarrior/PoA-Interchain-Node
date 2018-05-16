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
