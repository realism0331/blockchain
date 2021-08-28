package blockchain

import(
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dbPath="./tmp/blocks_%s"//changed at part10
	//dbFile="./tmp/blocks/MANIFEST"
	genesisData="First Transaction from Genesis"
)

type BlockChain struct{
	//Blocks []*Block//Capitalize the first letter is public
	LastHash []byte
	Database *badger.DB
}



/*func InitBlockChain() *BlockChain{
	//return &BlockChain{[]*Block{Genesis()}}//I don't know how it works?
	var lastHash []byte

	opts:=badger.DefaultOptions
	opts.Dir=dbPath
	opts.ValueDir=dbPath

	db,err:=badger.Open(opts)
	Handle(err)
	err=db.Update(func(txn *badger.Txn) error {
		if _,err:=txn.Get([]byte("lh"));err==badger.ErrKeyNotFound{
			fmt.Println("No existing blockchain found")
			genesis:=Genesis()
			fmt.Println("Genesis proved")
			err=txn.Set(genesis.Hash,genesis.Serialize())
			Handle(err)
			err=txn.Set([]byte("lh"),genesis.Hash)
			lastHash=genesis.Hash
			return err
		}else{
			item,err:=txn.Get([]byte("lh"))//lh may represent last-hash.this key's value is last block's hash
			Handle(err)
			lastHash,err=item.Value()
			return err
		}
	})
	Handle(err)
	blockchain :=BlockChain{lastHash,db}
	return &blockchain
}*/

func InitBlockChain(address,nodeId string) *BlockChain{
	//return &BlockChain{[]*Block{Genesis()}}//I don't know how it works?
	var lastHash []byte
	path:=fmt.Sprintf(dbPath,nodeId)

	if DBexists(path){
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts:=badger.DefaultOptions
	opts.Dir=path
	opts.ValueDir=path

	db,err:=openDB(path,opts)
	Handle(err)
	err=db.Update(func(txn *badger.Txn) error {
		cbtx:=CoinbaseTx(address,genesisData)
		genesis:=Genesis(cbtx)
		fmt.Println("Genesis created")
		err=txn.Set(genesis.Hash,genesis.Serialize())
		Handle(err)
		err=txn.Set([]byte("lh"),genesis.Hash)
		lastHash=genesis.Hash
		return err
	})
	Handle(err)
	blockchain :=BlockChain{lastHash,db}
	return &blockchain
}

func ContinueBlockChain(nodeId string) *BlockChain{//Open the database and return the last block'hash of the blockchain;the par before is address
	//if DBexists()==false{
	//	fmt.Println("No existing blockchain found, create one!")
	//	runtime.Goexit()
	//}
	path:=fmt.Sprintf(dbPath,nodeId)
	if DBexists(path)==false{//every node has its particular DB
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}


	var lastHash []byte

	opts:=badger.DefaultOptions
	opts.Dir=dbPath
	opts.ValueDir=dbPath

	db,err:=openDB(path,opts)//Open returns a new DB object
	Handle(err)

	err=db.Update(func(txn *badger.Txn) error {
		item,err:=txn.Get([]byte("lh"))
		Handle(err)
		lastHash,err=item.Value()
		return err
	})
	chain:=BlockChain{lastHash,db}
	return &chain
}

/*func(chain *BlockChain) AddBlock(data string){
	//prevBlock:=chain.Blocks[len(chain.Blocks)-1]
	//new:=CreateBlock(data,prevBlock.Hash)
	//chain.Blocks=append(chain.Blocks,new)
	var lastHash []byte

	err :=chain.Database.View(func(txn *badger.Txn) error {
		item,err:=txn.Get([]byte("lh"))
		Handle(err)
		lastHash,err=item.Value()
		return err
	})
	Handle(err)

	newBlock :=CreateBlock(data,lastHash)

	err =chain.Database.Update(func(txn *badger.Txn) error {
		err:=txn.Set(newBlock.Hash,newBlock.Serialize())
		Handle(err)
		err=txn.Set([]byte("lh"),newBlock.Hash)
		chain.LastHash=newBlock.Hash
		return err
	})
	Handle(err)
}*/

func (chain *BlockChain)GetBlockHashes()[][]byte{
	var blocks [][]byte
	iter:=chain.Iterator()
	for{
		block:=iter.Next()
		blocks=append(blocks,block.Hash)
		if len(block.PrevHash)==0{
			break
		}
	}
	return blocks
}

func(chain *BlockChain)GetBestHeight()int{
	var lastBlock Block
	err:=chain.Database.View(func(txn *badger.Txn) error {
		item,err:=txn.Get([]byte("lh"))
		Handle(err)
		lastHash,_:=item.Value()
		item,err =txn.Get(lastHash)
		Handle(err)
		lastBlockData,_:=item.Value()
		lastBlock=*Deserialize(lastBlockData)
		return nil
	})
	Handle(err)
	return lastBlock.Height
}

func (chain *BlockChain)GetBlock(blockHash []byte)(Block,error)  {
	var block Block
	err:=chain.Database.View(func(txn *badger.Txn) error {
		if item,err:=txn.Get(blockHash);err!=nil{
			return  errors.New("Block is not found")
		}else {
			blockData,_:=item.Value()
			block=*Deserialize(blockData)
		}
		return nil
	})
	if err!=nil{
		return block,err
	}
	return block,nil
}

func(chain *BlockChain) MineBlock(transaction []*Transaction) *Block{
	//prevBlock:=chain.Blocks[len(chain.Blocks)-1]
	//new:=CreateBlock(data,prevBlock.Hash)
	//chain.Blocks=append(chain.Blocks,new)
	var lastHash []byte
	var lastHeight int//added at part10

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()
		item,err =txn.Get(lastHash)
		Handle(err)
		lastBlockData,_:=item.Value()

		lastBlock:=Deserialize(lastBlockData)
		lastHeight =lastBlock.Height

		return err
	})
	Handle(err)

	newBlock := CreateBlock(transaction,lastHash,lastHeight)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
	return newBlock
}

func (chain *BlockChain)AddBlock(block *Block){
	err:=chain.Database.Update(func(txn *badger.Txn) error {
		if _,err:=txn.Get(block.Hash);err==nil{
			return  nil
		}
		blockData :=block.Serialize()
		err:=txn.Set(block.Hash,blockData)
		Handle(err)

		item,err:=txn.Get([]byte("lh"))//lh -> lastHash ->lastBlockData ->lastBlock
		Handle(err)
		lastHash,_:=item.Value()
		item,err=txn.Get(lastHash)
		Handle(err)
		lastBlockData,_:=item.Value()
		lastBlock:=Deserialize(lastBlockData)

		if block.Height>lastBlock.Height{
			err:=txn.Set([]byte("lh"),block.Hash)
			Handle(err)
			chain.LastHash=block.Hash
		}

		return nil
	})
	Handle(err)
}



func DBexists(path string) bool{//Determine whether a blockchain exists based on the database file, not the genesis block's name
	if _,err:=os.Stat(path+"/MANIFEST");os.IsNotExist(err){
		return false
	}
	return true
}

func retry(dir string,originalOpts badger.Options) (*badger.DB,error) {
	lockPath:=filepath.Join(dir,"LOCK")
	if err:=os.Remove(lockPath);err!=nil{
		return nil, fmt.Errorf(`removing "LOCK":%s`,err)
	}
	retryOpts:=originalOpts
	retryOpts.Truncate=true//DB path's "LOCK" is cut
	db,err:=badger.Open(retryOpts)
	return db,err

}

func openDB(dir string,opts badger.Options)(*badger.DB,error){
	if db,err:=badger.Open(opts);err!=nil{
		if strings.Contains(err.Error(),"LOCk"){//the lockfile represent there is another process is accessing the database
			if db,err:=retry(dir,opts);err==nil{
				log.Println("database unlocked,value log truncated")//truncated means cut
				return db,nil
			}
			log.Println("could not unlock database:",err)
		}
		return nil,err
	}else{
		return db,nil
	}
}
/*func (chain *BlockChain)FindUnspentTransactions(pubKeyHash []byte)[]Transaction  {
	var unspentTxs []Transaction
	spentTXOs :=make(map[string][]int)//make:to build a slice, txID  |  outIdx

	iter:=chain.Iterator()

	for {
		block :=iter.Next()
		for _,tx:=range block.Transactions{
			txID :=hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx,out:=range tx.Outputs{//to find unspend tx
				if spentTXOs[txID]!=nil{
					for _,spentOut:=range spentTXOs[txID] {
						if spentOut==outIdx{//compare the sequence number of Outputs
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash){// Except input, others access output also need validation
					unspentTxs=append(unspentTxs,*tx)
				}
			}
		if tx.IsCoinbase()==false{
			for _,in:=range tx.Inputs{
				if in.Userskey(pubKeyHash){////Determines whether the input has permission
					inTxID:=hex.EncodeToString(in.ID)
					spentTXOs[inTxID]=append(spentTXOs[inTxID],in.Out)
				}
			}
		}
		}
		if len(block.PreHash)==0{
			break
		}
	}
	return unspentTxs
}*/

/*func (chain *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput{// every unspend transaction have several unspend transaction outputs
	var UTXOs []TxOutput
	unspentTransactions :=chain.FindUnspentTransactions(pubKeyHash)
	for _,tx:=range unspentTransactions{
		for _,out :=range tx.Outputs{
			if out.IsLockedWithKey(pubKeyHash){
				UTXOs=append(UTXOs,out)
			}
		}
	}
	return UTXOs
}*/

func (chain *BlockChain) FindUTXO() map[string]TxOutputs { // every unspend transaction have several unspend transaction outputs  | also find spentTXOs along UTXO
	UTXO:=make(map[string]TxOutputs)//txID   |transaction
	spentTXOs:=make(map[string][]int) //txID  |   outIdx

	iter:=chain.Iterator()
	for{
		block:=iter.Next()

		for _,tx:=range block.Transactions{
			txID:=hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx,out:=range tx.Outputs{
				if spentTXOs[txID]!=nil{
					for _,spentout:=range spentTXOs[txID]{
						if spentout==outIdx{
							continue Outputs
						}
					}
				}
				outs:=UTXO[txID]
				outs.Outputs=append(outs.Outputs,out)
				UTXO[txID]=outs
			}
			if tx.IsCoinbase()==false{
				for _,in:=range tx.Inputs{
					inTxID:=hex.EncodeToString(in.ID)
					spentTXOs[inTxID]=append(spentTXOs[inTxID],in.Out)
				}
			}
		}
		if len(block.PrevHash)==0{
			break
		}
	}
	return UTXO
}

/*func (chain *BlockChain)FindSpendableOutputs(pubKeyHash []byte,amount int) (int,map[string][]int){// to create the general tx & make sure the address remains enough tokens;the amount is represent the tokens we want to send
	unspentOuts :=make(map[string][]int)
	unspentTxs:=chain.FindUnspentTransactions(pubKeyHash)
	accumulated:=0

	Work:
		for _,tx:=range unspentTxs{
			txID :=hex.EncodeToString(tx.ID)
			for outIdx,out:=range tx.Outputs{
				if out.IsLockedWithKey(pubKeyHash) && accumulated<amount{
					accumulated+=out.Value
					unspentOuts[txID]=append(unspentOuts[txID],outIdx)

					if accumulated >=amount{
						break Work
					}
				}
			}
		}
	return accumulated,unspentOuts
}*/

func (bc *BlockChain)FindTransaction(ID []byte)(Transaction,error)  {
	iter:=bc.Iterator()
	for{
		block :=iter.Next()
		for _,tx:=range block.Transactions{
			if bytes.Compare(tx.ID,ID)==0{
				return *tx,nil
			}
		}

		if len(block.PrevHash)==0{
			break
		}
	}
	return Transaction{},errors.New("Transaction does not exist")
}

func (bc *BlockChain)SignTransaction(tx *Transaction,privKey ecdsa.PrivateKey)  {//Create a transaction for each set of inputs
	preTxs:=make(map[string]Transaction)//Tx.ID

	for _,in:=range tx.Inputs{
		preTx,err:=bc.FindTransaction(in.ID)
		Handle(err)
		preTxs[hex.EncodeToString(preTx.ID)]=preTx
	}
	tx.Sign(privKey,preTxs)
}

func (bc *BlockChain)VerifyTransaction(tx *Transaction)  bool{
	preTxs:=make(map[string]Transaction)//Tx.ID
	if tx.IsCoinbase() {
		return true
	}

	for _,in:=range tx.Inputs{
		preTx,err:=bc.FindTransaction(in.ID)
		Handle(err)
		preTxs[hex.EncodeToString(preTx.ID)]=preTx
	}
	return tx.Verify(preTxs)
}


