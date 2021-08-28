package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/dgraph-io/badger"
	"log"
)

var (
	utxoPrefix=[]byte("utxo-")
	prefixLength =len(utxoPrefix)
)

type UTXOSet struct {
	BlockChain *BlockChain
}

func (u UTXOSet)Reindex()  {//Replace unspent outputs' keys in the database   | eg:utxo-key
	db:=u.BlockChain.Database
	u.DeleteByPrefix(utxoPrefix)
	UTXO:=u.BlockChain.FindUTXO()
	
	err:=db.Update(func(txn *badger.Txn) error {
		for txId,outs:=range UTXO {
			key,err:=hex.DecodeString(txId)
			Handle(err)
			key =append(utxoPrefix,key...)
			
			err=txn.Set(key,outs.Serialize())
			Handle(err)
		}
		return nil
	})
	Handle(err)
}

func (u *UTXOSet)Update(block *Block)  {//relate to the send function
	db:=u.BlockChain.Database
	
	err:=db.Update(func(txn *badger.Txn) error {
		for _,tx:=range block.Transactions{
			if tx.IsCoinbase()==false{
				for _,in:=range tx.Inputs{
					updatedOuts:=TxOutputs{}//Store outputs which are used by input 
					inID:=append(utxoPrefix,in.ID...)//The last TX is already have the utxoprefix going through the function reindex
					item,err:=txn.Get(inID)
					Handle(err)
					v,err:=item.Value()
					Handle(err)
					
					outs:=DeserializeOutputs(v)
					
					for outIdx,out:=range outs.Outputs{//the former block
						if outIdx!=in.Out{//Identifies this output in the transaction is UTXO
							updatedOuts.Outputs=append(updatedOuts.Outputs,out)
						}
					}
					if len(updatedOuts.Outputs)==0{//By this, you can determine which outputs are not reused as inputs
						if err:=txn.Delete(inID);err!=nil{
							log.Panic(err)
						}
					}else {
						if err:=txn.Set(inID,updatedOuts.Serialize());err!=nil{//why? the value should be txs' []byte not the outputs' []byte
							log.Panic(err)
						}
					}
				}
			}

			//just for coinbase
			newOutputs:=TxOutputs{}
			for _,out:=range tx.Outputs{// the current block
				newOutputs.Outputs=append(newOutputs.Outputs,out)
			}
			txID:=append(utxoPrefix,tx.ID...)
			if err:=txn.Set(txID,newOutputs.Serialize());err !=nil{
				log.Panic(err)
			}
		}
		return nil
	})
	Handle(err)
}

func (u *UTXOSet)DeleteByPrefix(prefix []byte)  {
	deleteKeys := func(KeysForDelete [][]byte) error{//to define a function
		if err:= u.BlockChain.Database.Update(func(txn *badger.Txn) error {
			for _,key:=range KeysForDelete{
				if err:=txn.Delete(key);err!=nil{
					return err
				} 
			}
			return nil
		});err !=nil{
			return err
		}
		return nil
	}
	collectSize:=100000
	u.BlockChain.Database.View(func(txn *badger.Txn) error {
		opts:=badger.DefaultIteratorOptions
		opts.PrefetchValues=false
		it:=txn.NewIterator(opts)
		defer it.Close()
		
		keysForDelete :=make([][]byte,0,collectSize)//the length of slice now is 0,but we have collectSize space in total
		KeysCollected :=0
		for it.Seek(prefix);it.ValidForPrefix(prefix);it.Next(){
			key:=it.Item().KeyCopy(nil)
			keysForDelete=append(keysForDelete,key)
			KeysCollected++
			if KeysCollected ==collectSize{
				if err:=deleteKeys(keysForDelete);err!=nil{//deleteKeys is a func that we define it ourselves
					log.Panic(err)
				}
				keysForDelete =make([][]byte,0,collectSize)
				KeysCollected =0
			}
		}
		if KeysCollected>0{
			if err:=deleteKeys(keysForDelete);err!=nil{
				log.Panic(err)
			}
		}
		return nil
	})
}

func (u UTXOSet)CountTransactions()int {//Count the amount of UTX
	db:=u.BlockChain.Database
	counter:=0
	err:=db.View(func(txn *badger.Txn) error {
		opts:=badger.DefaultIteratorOptions

		it:=txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix);it.ValidForPrefix(utxoPrefix);it.Next(){
			counter++
		}
		return nil
	})
	Handle(err)
	return counter
}

func (u UTXOSet)FindUnspentTransactions(pubKeyHash []byte)  []TxOutput{//relate to the getbalance func
	var UTXOs []TxOutput
	db:=u.BlockChain.Database
	err:=db.View(func(txn *badger.Txn) error {
		opts:=badger.DefaultIteratorOptions

		it:=txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix);it.ValidForPrefix(utxoPrefix);it.Next(){
			item:=it.Item()
			v,err:=item.Value()
			Handle(err)
			outs:=DeserializeOutputs(v)
			for _,out:=range outs.Outputs{
				if out.IsLockedWithKey(pubKeyHash){
					UTXOs=append(UTXOs,out)
				}
			}
		}
		return nil
	})
	Handle(err)
	return UTXOs
}

func (u UTXOSet)FindSpendableOutputs(pubKeyHash []byte,amount int) (int,map[string][]int){//relate to the send func; to create the general tx & make sure the address remains enough tokens;the amount is representing the tokens we want to send
	unspentOuts :=make(map[string][]int) //txID  |outIdx
	//unspentTxs:=chain.FindUnspentTransactions(pubKeyHash)
	accumulated:=0
	db:=u.BlockChain.Database

	err:=db.View(func(txn *badger.Txn) error {
		opts:=badger.DefaultIteratorOptions

		it:=txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix);it.ValidForPrefix(utxoPrefix);it.Next(){
			item:=it.Item()
			k:=item.Key()
			v,err:=item.Value()
			Handle(err)
			k=bytes.TrimPrefix(k,utxoPrefix)//attention
			txID:=hex.EncodeToString(k)
			outs:=DeserializeOutputs(v)
			for outIdx,out:=range outs.Outputs{
				if out.IsLockedWithKey(pubKeyHash) && accumulated<amount{
					accumulated+=out.Value
					unspentOuts[txID]=append(unspentOuts[txID],outIdx)
				}
			}
		}
		return nil
	})
	Handle(err)
	return accumulated,unspentOuts
}