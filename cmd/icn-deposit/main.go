// Utility to perform a transfer to another address on the other chain using the wallet

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"

	"github.com/WeTrustPlatform/interchain-node/bind/mainchain"
	"github.com/WeTrustPlatform/interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type depositable interface {
	Deposit(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error)
}

func main() {
	// Command line flags
	mainChainPtr := flag.Bool("mainchain", false, "Target the main chain wallet")
	sideChainPtr := flag.Bool("sidechain", false, "Target the side chain wallet")
	keyJSONPath := flag.String("keyjson", "", "Path to the JSON private key file of the user")
	password := flag.String("password", "", "Passphrase needed to unlock the user's JSON key")
	endpoint := flag.String("endpoint", "", "URL or path of the origin chain endpoint")
	walletStr := flag.String("wallet", "", "Ethereum address of the multisig wallet on the origin chain")
	receiver := flag.String("receiver", "", "Ethereum address of the receiver on the target chain")
	value := flag.String("value", "", "Value (wei) to transfer to the receiver")

	flag.Parse()

	// Connect to both chains
	client, _ := ethclient.Dial(*endpoint)

	walletAddress := common.HexToAddress(*walletStr)

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(*keyJSONPath)
	if err != nil {
		log.Fatalf("Key json read error: %v", err)
	}

	// Create a transactor
	auth, err := bind.NewTransactor(strings.NewReader(string(keyJSON[:])), *password)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	// Set the amount to be transfered
	v := new(big.Int)
	v, ok := v.SetString(*value, 10)
	if !ok {
		fmt.Println("SetString: error")
		return
	}
	auth.Value = v

	// Attach the wallet and submit the transaction
	var wallet depositable
	if *mainChainPtr {
		wallet, err = mainchain.NewMainChain(walletAddress, client)
	} else if *sideChainPtr {
		wallet, err = sidechain.NewSideChain(walletAddress, client)
	}

	if err != nil {
		log.Fatal("Couldn't instanciate the contract:", err)
	}

	wtx, err := wallet.Deposit(auth, common.HexToAddress(*receiver))
	if err != nil {
		log.Printf("SubmitTransaction error: %v", err)
	}
	log.Printf("Transaction sent: %v", wtx)
}
