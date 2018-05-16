# this makefile is setting up the development and manual testing environment, not building the interchain node
.DEFAULT_GOAL := run_sidechain

# generate the go bindings for the multisig wallets
bind/mainchain/main.go bind/sidechain/main.go:
	mkdir -p bind/mainchain/
	mkdir -p bind/sidechain/
	abigen --sol interchain-node-contracts/contracts/MainChain.sol --pkg mainchain --out bind/mainchain/main.go
	abigen --sol interchain-node-contracts/contracts/SideChain.sol --pkg sidechain --out bind/sidechain/main.go

# we use the same passphrase for every account in the dev env
password.txt:
	echo "dummy" > password.txt

# create mainchain accounts
mainchain/miner: password.txt
	mkdir -p mainchain
	geth --datadir ./mainchain --password ./password.txt account new | sed "s/Address: {\(.*\)}/\1/g" > mainchain/miner

# create sidechain accounts
sidechain/sealer: password.txt
	mkdir -p sidechain
	geth --datadir ./sidechain --password ./password.txt account new | sed "s/Address: {\(.*\)}/\1/g" > sidechain/sealer
	geth --datadir ./sidechain --password ./password.txt account new | sed "s/Address: {\(.*\)}/\1/g" > sidechain/tester

# create sidechain sealer 2 account
sidechain2/sealer: password.txt
	mkdir -p sidechain2
	geth --datadir ./sidechain2 --password ./password.txt account new | sed "s/Address: {\(.*\)}/\1/g" > sidechain2/sealer

# compute the future address of the mainchain wallet
mainchain/wallet: mainchain/miner
	eth-contract-address -address=`cat mainchain/miner` -nonce=0 > mainchain/wallet

# compute the future address of the sidechain wallet
sidechain/wallet: sidechain/sealer
	eth-contract-address -address=`cat sidechain/sealer` -nonce=0 > sidechain/wallet

# generate the sidechain genesis block from template
sidechain/genesis.json: sidechain/sealer sidechain2/sealer sidechain/wallet
	cat genesis/sidechain.json \
	| sed "s/@@SEALER1@@/`cat sidechain/sealer`/g" \
	| sed "s/@@SEALER2@@/`cat sidechain2/sealer`/g" \
	| sed "s/@@TESTER@@/`cat sidechain/tester`/g" \
	| sed "s/@@WALLET@@/`cat sidechain/wallet`/g" > sidechain/genesis.json

# generate the mainchain genesis block from template
mainchain/genesis.json: mainchain/miner sidechain/sealer sidechain2/sealer mainchain/wallet
	cat genesis/mainchain.json \
	| sed "s/@@MINER@@/`cat mainchain/miner`/g" \
	| sed "s/@@SEALER1@@/`cat sidechain/sealer`/g" \
	| sed "s/@@SEALER2@@/`cat sidechain2/sealer`/g" \
	| sed "s/@@WALLET@@/`cat mainchain/wallet`/g" > mainchain/genesis.json

# initialize the mainchain
mainchain/geth/nodekey: mainchain/genesis.json
	geth --nodiscover --datadir ./mainchain init mainchain/genesis.json

# initialize the first sidechain node
sidechain/geth/nodekey: sidechain/genesis.json
	geth --nodiscover --datadir ./sidechain init sidechain/genesis.json

# initialize the second sidechain node
sidechain2/geth/nodekey: sidechain/genesis.json
	geth --nodiscover --datadir ./sidechain2 init sidechain/genesis.json

# run the mainchain node
run_mainchain: mainchain/geth/nodekey
	geth --nodiscover --syncmode "full" --networkid 9007 --datadir ./mainchain --password ./password.txt --unlock `cat mainchain/miner` --mine --port 30307

# run the first sidechain node
run_sidechain: sidechain/geth/nodekey
	geth --nodiscover --syncmode "full" --networkid 9008 --datadir ./sidechain --identity "Sealer 1" --password ./password.txt --unlock `cat sidechain/sealer` --mine --etherbase `cat sidechain/sealer` --port 30308

# get the enode of the first sidechain node, needed for peer discovery
sidechain/enode sidechain2/static-nodes.json:
	echo "enode://$$(bootnode -nodekeyhex `cat ./sidechain/geth/nodekey` -writeaddress)@127.0.0.1:30308" > ./sidechain/enode
	echo "[ \"enode://$$(bootnode -nodekeyhex `cat ./sidechain/geth/nodekey` -writeaddress)@127.0.0.1:30308\" ]" > ./sidechain2/static-nodes.json

# run the second sidechain node
run_sidechain2: sidechain2/geth/nodekey sidechain/enode sidechain2/static-nodes.json
	geth --nodiscover --syncmode "full" --networkid 9008 --datadir ./sidechain2 --identity "Sealer 2" --password ./password.txt --unlock `cat sidechain2/sealer` --mine --etherbase `cat sidechain2/sealer` --port 30309 --bootnodes `cat sidechain/enode`

# deploy the multisig wallet on the mainchain
mainchain_wallet: bind/mainchain/main.go bind/sidechain/main.go sidechain/sealer sidechain2/sealer
	go run cmd/icn-deploy/main.go -mainchain -rpc=mainchain/geth.ipc -keyjson=mainchain/keystore/`ls -1 mainchain/keystore | head -n 1` -addresses="`cat sidechain/sealer`,`cat sidechain2/sealer`" -required=2 -password="dummy"

# deploy the multisig wallet on the sidechain
sidechain_wallet: bind/mainchain/main.go bind/sidechain/main.go sidechain/sealer sidechain2/sealer
	go run cmd/icn-deploy/main.go -sidechain -rpc=sidechain/geth.ipc -keyjson=sidechain/keystore/`ls -1 sidechain/keystore | head -n 1` -addresses="`cat sidechain/sealer`,`cat sidechain2/sealer`" -required=2 -password="dummy"

clean:
	rm -rf mainchain sidechain sidechain2 password.txt contracts/*.bin contracts/*.abi bind/*