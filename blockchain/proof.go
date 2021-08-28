package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

//Take the data from the block

//Create a counter(nounce) which starts at 0

//Create a  hash of the data plus the counter

//Check the hash to see if it meets a set of requirements

//Requirements:
//The First few bytes must contain 0s

const Difficulty = 18// This is static,but in real institution is dynamic

type ProofOfWork struct{
	Block *Block
	target *big.Int
}

func NewProof(b *Block) *ProofOfWork{
	target :=big.NewInt(1)
	target.Lsh(target,uint(256-Difficulty))//shift left

	pow:=&ProofOfWork{b,target}
	return pow
}

func(pow *ProofOfWork) InitData(nonce int)[] byte{
	data :=bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64((Difficulty))),
		},
		[]byte{},
		)
	return data
}

func(pow *ProofOfWork) Run()(int ,[]byte){
	var intHash big.Int
	var hash [32]byte
	nonce:=0

	for nonce<math.MaxInt64{
		data:=pow.InitData(nonce)
		hash =sha256.Sum256(data)

		fmt.Printf("\r%x",hash)

		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.target)==-1{//former is smaller
			break
		}else{
			nonce++
		}
	}
	//fmt.Printf("Target:\r%x",pow.target)
	fmt.Println()
	return nonce,hash[:]
}

func (pow *ProofOfWork) Validate() bool{
	var intHash big.Int
	data :=pow.InitData(pow.Block.Nonce)
	hash:=sha256.Sum256(data)//return Checksum of data
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.target)==-1//when I change the Difficulty,the result will become false,because the nonce isn't change along the difficulty,
}

func ToHex(num int64)[]byte{
	buff:=new(bytes.Buffer)
	err:=binary.Write(buff,binary.BigEndian,num)
	if err !=nil{
		log.Panic(err)
	}
	return buff.Bytes()
}