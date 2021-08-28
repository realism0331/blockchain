package cli

import (
	"flag"
	"fmt"
	"github.com/matilda/golang-blockchain/blockchain"
	"github.com/matilda/golang-blockchain/network"
	"github.com/matilda/golang-blockchain/wallet"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{
	blockchain  *blockchain.BlockChain//package blockchain's blockchain.go's BlockChain Struct
}
func (cli *CommandLine) printUsage(){
	fmt.Println("Usage:")
	fmt.Println("getbalance -address ADDRESS -get the balance for address")
	fmt.Println("createblockchain -address ADDRESS creates a blockChain")
	fmt.Println("printchain -Prints the blocks int the chain")
	fmt.Println("send -from FROM -to TO -amount AMOUNT -mine -Send amount from FROM to TO")
	fmt.Println("createwallet -Creates a new Wallet")
	fmt.Println("listaddresses - List the addresses in our wallet file")
	fmt.Println("reindexutxo -Rebuilds the UTXO setjuh")
	fmt.Println("countutxo - Get the amount of utxos")
	fmt.Println("startnode -miner ADDRESS -start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func(cli *CommandLine) validateArgs(){
	if len(os.Args)<2{
		cli.printUsage()
		runtime.Goexit()
	}
}

/*func (cli *CommandLine) addBlock(data string)  {
	cli.blockchain.AddBlock(data)
	fmt.Println("Added Block")
}*/

func (cli *CommandLine)StartNode(nodeID,minerAddress string){
	fmt.Printf("Starting Node %s\n",nodeID)

	if len(minerAddress)>0{
		if wallet.ValidateAddress(minerAddress){
			fmt.Printf("Mining is on.Address to receive rewards:%s\n",minerAddress)
		}else{
			log.Panic("Wrong miner address")
		}
	}
	network.StartServer(nodeID,minerAddress)
}

func (cli *CommandLine)reindexUTXO(nodeID string)  {
	chain :=blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	UTXOSet:=blockchain.UTXOSet{chain}
	UTXOSet.Reindex()

	count:=UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n",count)
}

func (cli *CommandLine)countUTXO(){
	chain :=blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	UTXOSet:=blockchain.UTXOSet{chain}

	count:=UTXOSet.CountTransactions()
	fmt.Printf("The amount of utxo is %d\n",count)
}

func (cli *CommandLine) printChain(nodeId string) {
	chain:=blockchain.ContinueBlockChain(nodeId)
	defer chain.Database.Close()
	iter :=chain.Iterator()
	for{
		block:=iter.Next()
		fmt.Printf("Previous Hash:%x\n",block.PrevHash)
		//fmt.Printf("Data in Block:%s\n",block.Data)
		fmt.Printf("Hash:%x\n",block.Hash)
		pow:=blockchain.NewProof(block)
		fmt.Printf("POW:%s\n",strconv.FormatBool(pow.Validate()))
		for _,tx:=range block.Transactions{
			fmt.Println(tx)
		}
		fmt.Printf("Nonce:%x\n",block.Nonce)
		fmt.Println()
		fmt.Printf("----------------------------\n")
		if len(block.PrevHash)==0{
			break
		}
	}
}

func (cli *CommandLine)createBlockChain(address ,nodeID string)  {
	if !wallet.ValidateAddress(address){
		log.Panicln("Address ia not valid")
	}
	chain:=blockchain.InitBlockChain(address,nodeID)
	defer chain.Database.Close()

	UTXOSet:=blockchain.UTXOSet{chain}
	UTXOSet.Reindex()

	fmt.Println("Finished!")

}
func (cli *CommandLine)getBalance(address string,nodeID string)  {//The condition that this function executes is reindexutxo function has been executed,so we modify the createBlockChain func
	if !wallet.ValidateAddress(address){
		log.Panicln("Address ia not valid")
	}
	chain:=blockchain.ContinueBlockChain(nodeID)
	UTXOSet:=blockchain.UTXOSet{chain}
	defer chain.Database.Close()
	balance:=0
	FullHash:=wallet.Base58Decode([]byte(address))
	pubKeyHash:=FullHash[1:len(FullHash)-4]
	UTXOs:=UTXOSet.FindUnspentTransactions(pubKeyHash)
	for _,out:=range UTXOs{
		balance +=out.Value
	}
	fmt.Printf("Balance of %s:%d\n",address,balance)
}

func (cli *CommandLine)send(from,to string,amount int,nodeID string,mineNow bool){
	if !wallet.ValidateAddress(from){
		log.Panicln("Address ia not valid")
	}
	if !wallet.ValidateAddress(to){
		log.Panicln("Address ia not valid")
	}
	chain:=blockchain.ContinueBlockChain(nodeID)
	UTXOSet:=blockchain.UTXOSet{chain}
	defer chain.Database.Close()

	wallets,err:=wallet.CreateWallets(nodeID)
	if err!=nil{
		log.Panic(err)
	}
	wallet:=wallets.GetWallet(from)

	tx:=blockchain.NewTransaction(&wallet,to,amount,&UTXOSet)
	if mineNow{
		cbTx:=blockchain.CoinbaseTx(from,"")
		txs:=[]*blockchain.Transaction{cbTx,tx}
		block:=chain.MineBlock(txs)
		UTXOSet.Update(block)
	}else{
		network.SendTx(network.KnownNodes[0],tx)
		fmt.Println("send tx")
	}
	//cbTx:=blockchain.CoinbaseTx(from,"")
	//block:=chain.AddBlock([]*blockchain.Transaction{cbTx,tx})
	//
	//UTXOSet.Update(block)
	fmt.Println("Success!")
}

func (cli *CommandLine)listAddresses(nodeID string)  {
	Wallets,_:=wallet.CreateWallets(nodeID)
	addresses:=Wallets.GetAllAddresses()

	for _,address:=range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine)createWallet(nodeID string)  {
	wallets,_:=wallet.CreateWallets(nodeID)//load DB
	address:=wallets.AddWallet()
	wallets.SaveFile(nodeID)//save the wallets to the DB

	fmt.Printf("New address is:%s\n",address)
}

func(cli *CommandLine) Run(){
	//cli.validateArgs()//just the amount of Args
	//addBlockCmd:=flag.NewFlagSet("add",flag.ExitOnError)//package flag can Process command line parameters
	//printChainCmd:=flag.NewFlagSet("print",flag.ExitOnError)//NewFlagSet() can declare a subcommand
	//addBlockData:=addBlockCmd.String("block","","Block data")
	cli.validateArgs()

	nodeID:=os.Getenv("NODE_ID")//Getting system variables
	if nodeID ==""{
		fmt.Printf("NODE_ID env is not set!")
		runtime.Goexit()
	}
	//action
	getBalanceCmd:=flag.NewFlagSet("getbalance",flag.ExitOnError)
	createBlockchainCmd:=flag.NewFlagSet("createblockchain",flag.ExitOnError)
	sendCmd:=flag.NewFlagSet("send",flag.ExitOnError)
	printChainCmd:=flag.NewFlagSet("printchain",flag.ExitOnError)
	createWalletCmd:=flag.NewFlagSet("createwallet",flag.ExitOnError)
	listAddressesCmd:=flag.NewFlagSet("listaddresses",flag.ExitOnError)
	reindexUTXOCmd:=flag.NewFlagSet("reindexUTXOCmd",flag.ExitOnError)
	countUTXOCmd:=flag.NewFlagSet("countUTXOCmd",flag.ExitOnError)
	startNodeCmd:=flag.NewFlagSet("startnode",flag.ExitOnError)

	//parameter
	getBalanceAddress:=getBalanceCmd.String("address","","The account address")
	createBlockchainAddress:=createBlockchainCmd.String("address","","The address of miner and coinbase address")
	sendFrom:=sendCmd.String("from","","Source wallet address")
	sendTo:=sendCmd.String("to","","Destination wallet address")
	sendAmount:=sendCmd.Int("amount",0,"Amount to send")
	sendMine:=sendCmd.Bool("mine",false,"Mine immediately on the same node")
	startNodeMiner:=startNodeCmd.String("miner","","Enable mining mode and send reword to the ADDRESS")

	switch os.Args[1]{
	//case "add":
	//	err:=addBlockCmd.Parse(os.Args[2:])//?????
	//	blockchain.Handle(err)
	//case "print":
	//	err:=printChainCmd.Parse(os.Args[2:])//?????
	//	blockchain.Handle(err)
	case "startnode":
		err:=startNodeCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "countutxo":
		err:=countUTXOCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "reindexutxo":
		err:=reindexUTXOCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "getbalance":
		err:=getBalanceCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "createblockchain":
		err:=createBlockchainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "printchain":
		err:=printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "send":
		err:=sendCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "createwallet":
		err:=createWalletCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "listaddresses":
		err:=listAddressesCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if startNodeCmd.Parsed(){
		nodeID:=os.Getenv("NODE_ID")
		if nodeID==""{
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID,*startNodeMiner)
	}

	if getBalanceCmd.Parsed(){//Parsed():Parse the command into the corresponding parameter
		if *getBalanceAddress==""{//Check parameters
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress,nodeID)
	}
	if createBlockchainCmd.Parsed(){//Parsed():Parse the command into the corresponding parameter
		if *createBlockchainAddress==""{
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress,nodeID)
	}
	if sendCmd.Parsed(){//Parsed():Parse the command into the corresponding parameter
		if *sendFrom=="" || *sendTo=="" ||*sendAmount<=0{
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom,*sendTo,*sendAmount,nodeID,*sendMine)
	}
	if createWalletCmd.Parsed(){
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed(){
		cli.listAddresses(nodeID)
	}

	if printChainCmd.Parsed(){
		cli.printChain(nodeID)
	}

	if reindexUTXOCmd.Parsed(){
		cli.reindexUTXO(nodeID)
	}

	if countUTXOCmd.Parsed(){
		cli.countUTXO()
	}
}

