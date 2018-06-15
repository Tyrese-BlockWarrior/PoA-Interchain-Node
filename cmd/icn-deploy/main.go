/*
Copyright (C) 2018 WeTrustPlatform

This file is part of poa-interchain-node.

poa-interchain-node is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

poa-interchain-node is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with poa-interchain-node.  If not, see <http://www.gnu.org/licenses/>.
*/

// Command line utility to deploy the multisig wallets
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"bufio"

	"github.com/WeTrustPlatform/poa-interchain-node/bind/mainchain"
	"github.com/WeTrustPlatform/poa-interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	MainChain   bool   `short:"m" long:"mainchain" description:"Deploy the main chain wallet"`
	SideChain   bool   `short:"s" long:"sidechain" description:"Deploy the side chain wallet"`
	KeyJSONPath string `short:"k" long:"keyjson" required:"true" description:"Path to the key json file of the transactor"`
	Password    string `short:"p" long:"password" description:"Password needed to unlock the keyjson"`
	Addresses   string `short:"a" long:"addresses" required:"true" description:"Comma separated list of the owners"`
	Required 		int64  `long:"required" default:"2" description:"Number of votes required for a transaction, must be inferior or equal to the number of owners"`
	RPC         string `long:"rpc" default:"http://127.0.0.1:8545" description:"Address of the node RPC endpoint, can be HTTP or IPC"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	// Prompt passphrase if not passed as a flag
	if opts.Password == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your passphrase: ")
		opts.Password, _ = reader.ReadString('\n')
		opts.Password = strings.TrimSuffix(opts.Password, "\n")
	}

	// Connect to the node
	conn, err := ethclient.Dial(opts.RPC)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(opts.KeyJSONPath)
	if err != nil {
		log.Fatal("Key json read error:", err)
	}

	// Create a transactor
	auth, err := bind.NewTransactor(strings.NewReader(string(keyJSON[:])), opts.Password)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	// Build the list of owners
	hexAddresses := strings.Split(opts.Addresses, ",")
	fmt.Println(hexAddresses)
	addresses := []common.Address{}
	for _, hexAddress := range hexAddresses {
		addresses = append(addresses, common.HexToAddress(hexAddress))
	}

	// Deploy the wallet
	if opts.MainChain {
		address, _, _, err := mainchain.DeployMainChain(auth, conn, addresses, uint8(opts.Required))
		if err != nil {
			log.Fatalf("Failed to deploy new mainchain wallet contract: %v", err)
		}
		fmt.Printf("%x\n", address)
	}

	if opts.SideChain {
		address, _, _, err := sidechain.DeploySideChain(auth, conn, addresses, uint8(opts.Required))
		if err != nil {
			log.Fatalf("Failed to deploy new sidechain wallet contract: %v", err)
		}
		fmt.Printf("%x\n", address)
	}
}
