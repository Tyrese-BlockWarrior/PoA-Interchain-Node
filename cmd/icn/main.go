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

package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/WeTrustPlatform/interchain-node"
	"github.com/WeTrustPlatform/interchain-node/bind/mainchain"
	"github.com/WeTrustPlatform/interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	MainChain         bool   `short:"m" long:"mainchain" required:"false" description:"Watch the main chain only"`
	SideChain         bool   `short:"s" long:"sidechain" required:"false" description:"Watch the side chain only"`
	KeyJSONPath       string `short:"k" long:"keyjson" required:"true" description:"Path to the JSON private key file of the sealer"`
	Password          string `short:"p" long:"password" required:"false" description:"Passphrase needed to unlock the sealer's JSON key"`
	MainChainEndpoint string `long:"mainchainendpoint" required:"true" description:"URL or path of the main chain endpoint"`
	SideChainEndpoint string `long:"sidechainendpoint" required:"true" description:"URL or path of the side chain endpoint"`
	MainChainWallet   string `long:"mainchainwallet" required:"true" description:"Ethereum address of the multisig wallet on the main chain"`
	SideChainWallet   string `long:"sidechainwallet" required:"true" description:"Ethereum address of the multisig wallet on the side chain"`
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
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
	}

	// Default behavior is to watch both chains
	if !opts.MainChain && !opts.SideChain {
		opts.SideChain = true
		opts.MainChain = true
	}

	// Connect to both chains
	mainChainClient, err := ethclient.Dial(opts.MainChainEndpoint)
	handleError(err)
	sideChainClient, err := ethclient.Dial(opts.SideChainEndpoint)
	handleError(err)

	sideChainWalletAddress := common.HexToAddress(opts.SideChainWallet)
	mainChainWalletAddress := common.HexToAddress(opts.MainChainWallet)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(opts.KeyJSONPath)
	handleError(err)

	// Create a transactor
	key, err := keystore.DecryptKey(keyJSON, opts.Password)
	handleError(err)
	auth := bind.NewKeyedTransactor(key.PrivateKey)

	// Attach the wallet
	sc, err := sidechain.NewSideChain(sideChainWalletAddress, sideChainClient)
	handleError(err)
	mc, err := mainchain.NewMainChain(mainChainWalletAddress, mainChainClient)
	handleError(err)

	var wg sync.WaitGroup

	if opts.MainChain {
		wg.Add(1)
		go func() {
			mci, _ := mc.FilterDeposit(&bind.FilterOpts{Start: 0, End: nil, Context: ctx}, []common.Address{}, []common.Address{})
			for mci.Next() {
				tx, err := sc.SubmitTransactionSC(auth, mci.Event.Raw.TxHash, mci.Event.To, mci.Event.Value, []byte{})
				log.Println("[mc2sc]", mci.Event.Raw.BlockNumber, tx, err)
			}
			wg.Done()
		}()
	}

	if opts.SideChain {
		wg.Add(1)
		go func() {
			sci, _ := sc.FilterDeposit(&bind.FilterOpts{Start: 0, End: nil, Context: ctx}, []common.Address{}, []common.Address{})
			for sci.Next() {
				tx, err := icn.SubmitSignatureMC(ctx, sideChainWalletAddress, auth, sc, sci.Event, key.PrivateKey)
				log.Println("[sc2mc]", sci.Event.Raw.BlockNumber, tx, err)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
