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

// Utility to perform a transfer to another address on the other chain using the wallet
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"bufio"

	"github.com/WeTrustPlatform/poa-interchain-node/bind/mainchain"
	"github.com/WeTrustPlatform/poa-interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	MainChain   bool   `short:"m" long:"mainchain" description:"Target the main chain wallet"`
	SideChain   bool   `short:"s" long:"sidechain" description:"Target the side chain wallet"`
	KeyJSONPath string `short:"k" long:"keyjson" required:"true" description:"Path to the JSON private key file of the sealer"`
	Password    string `short:"p" long:"password"  description:"Passphrase needed to unlock the sealer's JSON key"`
	Endpoint    string `short:"e" long:"endpoint" required:"true" description:"URL or path of the origin chain endpoint"`
	Wallet      string `short:"w" long:"wallet" required:"true"  description:"Ethereum address of the multisig wallet on the origin chain"`
	Receiver    string `short:"r" long:"receiver" required:"true" description:"Ethereum address of the receiver on the target chain"`
	Value       string `short:"v" long:"value" required:"true" description:"Value (wei) to transfer to the receiver"`
}

type depositable interface {
	Deposit(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error)
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

	// Connect to both chains
	client, _ := ethclient.Dial(opts.Endpoint)

	walletAddress := common.HexToAddress(opts.Wallet)

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(opts.KeyJSONPath)
	if err != nil {
		log.Fatalf("Key json read error: %v", err)
	}

	// Create a transactor
	auth, err := bind.NewTransactor(strings.NewReader(string(keyJSON[:])), opts.Password)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	// Set the amount to be transfered
	v := new(big.Int)
	v, ok := v.SetString(opts.Value, 10)
	if !ok {
		fmt.Println("SetString: error")
		return
	}
	auth.Value = v

	// Attach the wallet and submit the transaction
	var wallet depositable
	if opts.MainChain {
		wallet, err = mainchain.NewMainChain(walletAddress, client)
	} else if opts.SideChain {
		wallet, err = sidechain.NewSideChain(walletAddress, client)
	} else {
		log.Printf("please specify which chain the wallet should be attached to in the command. Using --mainchain or --sidechain")
	}

	if err != nil {
		log.Fatal("Couldn't instanciate the contract:", err)
	}

	wtx, err := wallet.Deposit(auth, common.HexToAddress(opts.Receiver))
	if err != nil {
		log.Printf("Deposit error: %v", err)
	}
	log.Printf("Transaction sent: %v", wtx.Hash().String())
}
