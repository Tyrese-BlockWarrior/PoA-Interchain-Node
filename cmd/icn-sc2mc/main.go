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
	"github.com/WeTrustPlatform/interchain-node/bind/sidechain"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func proceedTransaction(
	ctx context.Context,
	sideChainWalletAddress common.Address,
	auth *bind.TransactOpts,
	sidechainClient *ethclient.Client,
	sc *sidechain.SideChain,
	tx *types.Transaction,
	key *keystore.Key,
) error {
	// Decode event logs
	abi, _ := abi.JSON(strings.NewReader(sidechain.SideChainABI))
	logs, err := icn.GetLogs(ctx, sidechainClient, tx)
	if err != nil {
		return err
	}
	deposit := icn.GetDepositEvent(abi, logs)

	if deposit == (icn.DepositEvent{}) {
		return errors.New("No deposit event")
	}

	log.Println("Mirroring transaction")

	// Submit the transaction
	var data []byte
	msgHash := icn.MsgHash(sideChainWalletAddress, tx.Hash(), deposit.Receiver, tx.Value(), data, 1)

	sig, err := crypto.Sign(msgHash.Bytes(), key.PrivateKey)
	if err != nil {
		return errors.New("Sign failed: " + err.Error())
	}

	r := common.BytesToHash(sig[0:32])
	s := common.BytesToHash(sig[32:64])
	v := uint8(sig[64:65][0] + 27)

	wtx, err := sc.SubmitSignatureMC(auth, tx.Hash(), deposit.Receiver, tx.Value(), data, v, r, s)
	if err != nil {
		return errors.New("SubmitSignatureMC failed: " + err.Error())
	}

	log.Printf("Transaction mirrored: %v", wtx)
	return nil
}

func main() {
	// Command line flags
	keyJSONPath := flag.String("keyjson", "", "Path to the JSON private key file of the sealer")
	password := flag.String("password", "", "Passphrase needed to unlock the sealer's JSON key")
	//mainChainEndpoint := flag.String("mainchainendpoint", "", "URL or path of the main chain endpoint")
	sideChainEndpoint := flag.String("sidechainendpoint", "", "URL or path of the side chain endpoint")
	//mainChainWallet := flag.String("mainchainwallet", "", "Ethereum address of the multisig wallet on the main chain")
	sideChainWallet := flag.String("sidechainwallet", "", "Ethereum address of the multisig wallet on the side chain")

	flag.Parse()

	// Connect to both chains
	//mainChainClient, _ := ethclient.Dial(*mainChainEndpoint)
	sideChainClient, _ := ethclient.Dial(*sideChainEndpoint)

	sideChainWalletAddress := common.HexToAddress(*sideChainWallet)
	//mainChainWalletAddress := common.HexToAddress(*mainChainWallet)

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
	latestBlock, err := sideChainClient.BlockByNumber(ctx, nil)
	if err != nil {
		log.Fatal("Can't get latest block:", err)
	}

	log.Printf("Latest block: %v", latestBlock.Number())

	// Loop over the blocks
	for i := big.NewInt(0); i.Cmp(latestBlock.Number()) <= 0; i.Add(i, big.NewInt(1)) {

		// Get the block details
		block, err := sideChainClient.BlockByNumber(ctx, i)
		if err != nil {
			log.Println("Can't get block:", err)
			continue
		}

		txs := block.Transactions()

		// Loop over the transactions
		for j, tx := range txs {
			to := tx.To()

			// If money is sent to the side chain wallet address, mirror the transaction on the main chain
			if to != nil && *to == sideChainWalletAddress {
				err := proceedTransaction(ctx, sideChainWalletAddress, auth, sideChainClient, sc, tx, key)
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("Transaction proceeded in block %v: %v\n", i, j)
				}
			}
		}

		//log.Println("Block proceeded:", i)
	}
}
