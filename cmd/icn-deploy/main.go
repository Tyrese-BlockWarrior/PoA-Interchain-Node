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

// Command line utility to deploy the multisig wallets
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/WeTrustPlatform/interchain-node/bind/mainchain"
	"github.com/WeTrustPlatform/interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	mainChainPtr := flag.Bool("mainchain", false, "Deploy the main chain wallet")
	sideChainPtr := flag.Bool("sidechain", false, "Deploy the side chain wallet")
	rpcPtr := flag.String("rpc", "http://127.0.0.1:8545", "Address of the node RPC endpoint, can be HTTP or IPC")
	keyJSONPtr := flag.String("keyjson", "", "Path to the key json file of the transactor")
	passwordPtr := flag.String("password", "", "Password needed to unlock the keyjson")
	addressesPtr := flag.String("addresses", "", "Comma separated list of the owners")
	requiredPtr := flag.Int64("required", 2, "Number of votes required for a transaction, must be inferior or equal to the number of owners")

	flag.Parse()

	// Connect to the node
	conn, err := ethclient.Dial(*rpcPtr)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(*keyJSONPtr)
	if err != nil {
		log.Fatal("Key json read error:", err)
	}

	// Create a transactor
	auth, err := bind.NewTransactor(strings.NewReader(string(keyJSON[:])), *passwordPtr)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	// Build the list of owners
	hexAddresses := strings.Split(*addressesPtr, ",")
	addresses := []common.Address{}
	for _, hexAddress := range hexAddresses {
		addresses = append(addresses, common.HexToAddress(hexAddress))
	}

	// Deploy the wallet
	if *mainChainPtr {
		address, _, _, err := mainchain.DeployMainChain(auth, conn, addresses, uint8(*requiredPtr))
		if err != nil {
			log.Fatalf("Failed to deploy new mainchain wallet contract: %v", err)
		}
		fmt.Printf("%x\n", address)
	}

	if *sideChainPtr {
		address, _, _, err := sidechain.DeploySideChain(auth, conn, addresses, uint8(*requiredPtr))
		if err != nil {
			log.Fatalf("Failed to deploy new sidechain wallet contract: %v", err)
		}
		fmt.Printf("%x\n", address)
	}
}
