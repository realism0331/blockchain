package main

import (
	"github.com/matilda/golang-blockchain/cli"
	"os"
)


func main()  {
	//initiate

	defer os.Exit(0)
	//chain :=blockchain.InitBlockChain("145XvF5t8KR26HquWtRxKZZyEEdkePqbHY")
	//defer chain.Database.Close()
	//cli:=cli.CommandLine{(chain)}
	cli:=cli.CommandLine{}
	cli.Run()
	//w:=wallet.MakeWallet()
	//w.Address()

	//mininggo
	/*chain.AddBlock("First Block after Genesis")-
	chain.AddBlock("Third Block after Genesis")*/

	//traverse
	/*for _,block :=range chain.Blocks{
		fmt.Printf("Previous Hash:%x\n",block.PreHash)
		fmt.Printf("Data in Block:%s\n",block.Data)
		fmt.Printf("Hash:%x\n",block.Hash)


		pow:=blockchain.NewProof(block)
		fmt.Printf("POW:%s\n",strconv.FormatBool(pow.Validate()))
		fmt.Printf("Nonce:%x\n",block.Nonce)
		fmt.Println()
		fmt.Printf("----------------------------\n")
	}*/
}
