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
