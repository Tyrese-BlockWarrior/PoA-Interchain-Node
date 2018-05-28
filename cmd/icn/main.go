package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/WeTrustPlatform/interchain-node"
	"github.com/WeTrustPlatform/interchain-node/bind/mainchain"
	"github.com/WeTrustPlatform/interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func handleError(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func main() {
	// Command line flags
	keyJSONPath := flag.String("keyjson", "", "Path to the JSON private key file of the sealer")
	password := flag.String("password", "", "Passphrase needed to unlock the sealer's JSON key")
	mainChainEndpoint := flag.String("mainchainendpoint", "", "URL or path of the main chain endpoint")
	sideChainEndpoint := flag.String("sidechainendpoint", "", "URL or path of the side chain endpoint")
	mainChainWallet := flag.String("mainchainwallet", "", "Ethereum address of the multisig wallet on the main chain")
	sideChainWallet := flag.String("sidechainwallet", "", "Ethereum address of the multisig wallet on the side chain")

	flag.Parse()

	// Connect to both chains
	mainChainClient, err := ethclient.Dial(*mainChainEndpoint)
	handleError(err)
	sideChainClient, err := ethclient.Dial(*sideChainEndpoint)
	handleError(err)

	sideChainWalletAddress := common.HexToAddress(*sideChainWallet)
	mainChainWalletAddress := common.HexToAddress(*mainChainWallet)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(*keyJSONPath)
	handleError(err)

	// Create a transactor
	key, err := keystore.DecryptKey(keyJSON, *password)
	handleError(err)
	auth := bind.NewKeyedTransactor(key.PrivateKey)

	// Attach the wallet
	sc, err := sidechain.NewSideChain(sideChainWalletAddress, sideChainClient)
	handleError(err)
	mc, err := mainchain.NewMainChain(mainChainWalletAddress, mainChainClient)
	handleError(err)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		mci, _ := mc.FilterDeposit(&bind.FilterOpts{Start: 0, End: nil, Context: ctx}, []common.Address{}, []common.Address{})
		for mci.Next() {
			tx, err := sc.SubmitTransactionSC(auth, mci.Event.Raw.TxHash, mci.Event.To, mci.Event.Value, []byte{})
			log.Println("[mc2sc]", mci.Event.Raw.BlockNumber, tx, err)
		}
		wg.Done()
	}()

	go func() {
		sci, _ := sc.FilterDeposit(&bind.FilterOpts{Start: 0, End: nil, Context: ctx}, []common.Address{}, []common.Address{})
		for sci.Next() {
			tx, err := icn.SubmitSignatureMC(ctx, sideChainWalletAddress, auth, sc, sci.Event, key.PrivateKey)
			log.Println("[sc2mc]", sci.Event.Raw.BlockNumber, tx, err)
		}
		wg.Done()
	}()

	wg.Wait()
}
