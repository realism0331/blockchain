package blockchain

import (
	"bytes"
	"encoding/gob"
	"github.com/matilda/golang-blockchain/wallet"
)

type TxInput struct {
	ID []byte //the id of reference of tx
	Out int // the id of UTXO of tx
	Signature []byte // unlock the locked script,may be the Sk of account
	Pubkey []byte //hasn't been hash
}

type TxOutputs struct {
	Outputs []TxOutput
}

type TxOutput struct {//UTXO,The index of out id to be the outID
	Value int  //the amount of tokens
	PubKeyHash []byte//scriptPubkey:to lock the script, if subsequent transaction want  to use this output as input, must use Sig unlock tht locked script
}

/*func (in *TxInput) CanUnlock(data string) bool{
	return in.Sig==data
}

func(out *TxOutput)CanBeUnlocked(data string)bool{
	return out.PubKey==data
}*/
func NewTXOutput(value int,address string) *TxOutput {
	txo:=&TxOutput{value,nil}
	txo.Lock([]byte(address))
	return txo
}

func (in *TxInput)Userskey(pubKeyHash []byte)  bool{//Determines whether the input has permission
	lockingHash:=wallet.PublicKeyHash(in.Pubkey)

	return bytes.Compare(lockingHash,pubKeyHash)==0
}

func(out *TxOutput)Lock(address []byte){
	FullHash :=wallet.Base58Decode(address)
	pubkeyHash:=FullHash[1:len(FullHash)-4]
	out.PubKeyHash=pubkeyHash
}

func (out *TxOutput)IsLockedWithKey(pubKeyHash []byte)bool  {// Except input, others access output also need validation
	return bytes.Compare(out.PubKeyHash,pubKeyHash)==0
}

func (outs TxOutputs)Serialize() []byte {
	var buffer bytes.Buffer
	encode:=gob.NewEncoder(&buffer)
	err:=encode.Encode(outs)
	Handle(err)
	return buffer.Bytes()
}

func DeserializeOutputs(data []byte)TxOutputs{
	var outputs TxOutputs
	decode:=gob.NewDecoder(bytes.NewReader(data))
	err:=decode.Decode(&outputs)
	Handle(err)
	return  outputs
}