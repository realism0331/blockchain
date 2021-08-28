package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)



type Block struct{
	Timestamp int64//added at part10
	Hash []byte
	//Data []byte
	Transactions []*Transaction//appended after the transaction.go built
	PrevHash []byte
	Nonce int
	Height int//added at part10
}
/*func (b *Block) DeriveHash(){
	info:=bytes.Join([][]byte{b.Data,b.PreHash},[]byte{},0)
	hash:=sha256.Sum256(info)
	b.Hash=hash[:]

}*/

/*func CreateBlock(data string,prevHash []byte) *Block{
	block:=&Block{[]byte{},[]byte(data),prevHash,0}//initial & type conversion
	pow:=NewProof(block)
	nonce,hash:=pow.Run()
	//fmt.Printf("\r%x",hash)
	block.Nonce=nonce
	block.Hash=hash[:]
	return block
}*/


func CreateBlock(txs []*Transaction,prevHash []byte,height int) *Block{
	block:=&Block{time.Now().Unix(),[]byte{},txs,prevHash,0,height}//initial & type conversion
	pow:=NewProof(block)
	nonce,hash:=pow.Run()
	//fmt.Printf("\r%x",hash)
	block.Nonce=nonce
	block.Hash=hash[:]
	return block
}

/*func Genesis() *Block{
	return CreateBlock("Genesis",[]byte{})
}*/

func Genesis(coinbase *Transaction) *Block{
	return CreateBlock([]*Transaction{coinbase},[]byte{},0)
}

func (b *Block)HashTransactions()[]byte{
	var txHashes [][]byte
	//var txHash [32]byte
	for _,tx:=range b.Transactions{
		txHashes=append(txHashes,tx.Serialize())
	}
	tree:=NewMkerkleTree(txHashes)
	//txHash=sha256.Sum256(bytes.Join(txHashes,[]byte{}))//why append a blank byte?Sum256 returns the SHA256 checksum of the data
	return tree.RootNode.Data
}



func(b *Block)Serialize() []byte{
	var res bytes.Buffer
	encoder :=gob.NewEncoder(&res)

	err:=encoder.Encode(b)

	Handle(err)
	return res.Bytes()
}

func Deserialize(data []byte) *Block{
	var block Block
	decoder :=gob.NewDecoder(bytes.NewReader(data))

	err:=decoder.Decode(&block)

	Handle(err)

	return &block
}

func Handle (err error)  {
	if err !=nil{
		log.Panic(err)
	}
}