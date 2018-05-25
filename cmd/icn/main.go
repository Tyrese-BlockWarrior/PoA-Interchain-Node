package main

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/WeTrustPlatform/interchain-node"
	"github.com/WeTrustPlatform/interchain-node/bind/mainchain"
	"github.com/WeTrustPlatform/interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func submitSignatureMC(
	ctx context.Context,
	sideChainWalletAddress common.Address,
	auth *bind.TransactOpts,
	sidechainClient *ethclient.Client,
	sc *sidechain.SideChain,
	di icn.DepositInfo,
	key *keystore.Key,
) (*types.Transaction, error) {
	// Create the message hash
	var data []byte

	msgHash := icn.MsgHash(sideChainWalletAddress, di.TxHash, di.Event.Receiver, di.Event.Value, data, 1)

	// Sign the message hash
	sig, err := crypto.Sign(msgHash.Bytes(), key.PrivateKey)
	if err != nil {
		return nil, errors.New("Sign failed: " + err.Error())
	}

	// Parse the signature
	v, r, s := icn.ParseSignature(sig)

	// Submit the signature
	return sc.SubmitSignatureMC(auth, di.TxHash, di.Event.Receiver, di.Event.Value, data, v, r, s)
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
	if err != nil {
		log.Fatalf(err.Error())
	}
	sideChainClient, err := ethclient.Dial(*sideChainEndpoint)
	if err != nil {
		log.Fatalf(err.Error())
	}

	sideChainWalletAddress := common.HexToAddress(*sideChainWallet)
	mainChainWalletAddress := common.HexToAddress(*mainChainWallet)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(*keyJSONPath)
	if err != nil {
		log.Fatalf("Key json read error: %v", err)
	}

	// Create a transactor
	key, err := keystore.DecryptKey(keyJSON, *password)
	if err != nil {
		log.Fatalf("Failed to decrypt key: %v", err)
	}
	auth := bind.NewKeyedTransactor(key.PrivateKey)

	// Attach the wallet
	sc, err := sidechain.NewSideChain(sideChainWalletAddress, sideChainClient)
	if err != nil {
		log.Fatal("Couldn't instanciate the contract:", err)
	}

	// Get the latest block
	mcLast, err := mainChainClient.BlockByNumber(ctx, nil)
	if err != nil {
		log.Fatal("Can't get latest block:", err)
	}
	scLast, err := sideChainClient.BlockByNumber(ctx, nil)
	if err != nil {
		log.Fatal("Can't get latest block:", err)
	}

	mcABI, _ := abi.JSON(strings.NewReader(mainchain.MainChainABI))
	scABI, _ := abi.JSON(strings.NewReader(sidechain.SideChainABI))

	mcDeposits := make(chan icn.DepositInfo)
	scDeposits := make(chan icn.DepositInfo)
	done := make(chan bool)

	go icn.FindDeposits(
		ctx,
		mainChainClient,
		mcABI,
		mcDeposits,
		done,
		big.NewInt(0),
		mcLast.Number(),
		mainChainWalletAddress)

	go icn.FindDeposits(
		ctx,
		sideChainClient,
		scABI,
		scDeposits,
		done,
		big.NewInt(0),
		scLast.Number(),
		sideChainWalletAddress)

	for n := 2; n > 0; {
		select {
		case d := <-mcDeposits:
			tx, err := sc.SubmitTransactionSC(auth, d.TxHash, d.Event.Receiver, d.Event.Value, []byte{})
			log.Println("[mc2sc]", tx, err)
		case d := <-scDeposits:
			tx, err := submitSignatureMC(ctx, sideChainWalletAddress, auth, sideChainClient, sc, d, key)
			log.Println("[sc2mc]", tx, err)
		case <-done:
			n--
		}
	}
}
