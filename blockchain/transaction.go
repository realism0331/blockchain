package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/matilda/golang-blockchain/wallet"
	"log"
	"math/big"
	"strings"
)

type Transaction struct {
	ID []byte
	Inputs []TxInput
	Outputs []TxOutput
}

func DeserializeTransaction(data []byte)Transaction{
	var transaction Transaction

	decoder:=gob.NewDecoder(bytes.NewReader(data))
	err:=decoder.Decode(&transaction)
	Handle(err)
	return transaction
}


func CoinbaseTx(to,data string) *Transaction{
	if data==""{
		randData :=make([]byte,24)
		_,err:=rand.Read(randData)
		if err!=nil{
			log.Panic(err)
		}
		data =fmt.Sprintf("%x",randData)
	}
	txin:=TxInput{[]byte{},-1,nil,[]byte(data)}
	txout:=NewTXOutput(20,to)

	tx:=Transaction{nil,[]TxInput{txin},[]TxOutput{*txout}}
	tx.ID=tx.Hash()
	return &tx
}

func NewTransaction(w *wallet.Wallet,to string,amount int,UTXO *UTXOSet,)  *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	//wallets,err:=wallet.CreateWallets(nodeID)
	//Handle(err)
	//w:=wallets.GetWallet(from)
	pubKeyHash:=wallet.PublicKeyHash(w.PublicKey)
	acc,validOutputs:=UTXO.FindSpendableOutputs(pubKeyHash,amount)//there is case where the sum of all spendableOutputs is less than the amount;validOutputs: TxID,UTXOId

	if acc<amount{
		log.Panicln("Error:not enough funds")
	}
	//if remains is enough,The output needs to be turned into the input for the next round of transactions
	for txid,outs:=range validOutputs{
		txID,err:=hex.DecodeString(txid)
		Handle(err)

		for _,out:=range outs{//outs is a set of UTXOId
			input :=TxInput{txID,out,nil,w.PublicKey}
			inputs =append(inputs,input)
		}
	}

	from:=fmt.Sprintf("%s",w.Address())
	outputs =append(outputs,*NewTXOutput(amount,to))
	if acc>amount{
		outputs=append(outputs,*NewTXOutput(acc-amount,from))
	}
	tx:=Transaction{nil,inputs,outputs}
	tx.ID=tx.Hash()
	//tx.SetID()
	UTXO.BlockChain.SignTransaction(&tx,w.PrivateKey)
	return &tx
}

func (tx Transaction)Serialize() []byte {//To convert struct tx to bytes tx
	var encoded bytes.Buffer

	enc:=gob.NewEncoder(&encoded)
	err:=enc.Encode(tx)
	if err !=nil{
		log.Panic(err)
	}
	return encoded.Bytes()
}

func (tx *Transaction)Hash() []byte{//Assign a value to the id of the transaction
	var hash [32]byte
	txCopy:= *tx
	txCopy.ID=[]byte{}

	hash=sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

/*func (tx *Transaction)SetID(){// equal to Hash()
	var encoded bytes.Buffer
	var hash [32]byte

	encode :=gob.NewEncoder(&encoded)
	err:=encode.Encode(tx)

	Handle(err)
	hash=sha256.Sum256(encoded.Bytes())
	tx.ID=hash[:]
}*/

func (tx *Transaction) IsCoinbase() bool{
	return len(tx.Inputs)==1&&len(tx.Inputs[0].ID)==0&&tx.Inputs[0].Out==-1
}

func (tx *Transaction)TrimmedCopy()Transaction  {
	var inputs []TxInput
	var outputs []TxOutput
	for _,in:=range tx.Inputs{
		inputs =append(inputs,TxInput{in.ID,in.Out,nil,nil})
	}

	for _,out:=range tx.Outputs{
		outputs =append(outputs,TxOutput{out.Value,out.PubKeyHash})
	}

	txCopy:=Transaction{tx.ID,inputs,outputs}

	return txCopy
}

func (tx *Transaction)Sign(privKey ecdsa.PrivateKey,prevTXs map[string]Transaction)  {
	if tx.IsCoinbase(){//coinbase needn't to be Signed
		return
	}
	for _,in :=range tx.Inputs{
		if prevTXs[hex.EncodeToString(in.ID)].ID==nil{
			log.Panic("ERROR:Previous transaction is not correct")
		}
	}
	txCopy:=tx.TrimmedCopy()
	for inId,in:=range txCopy.Inputs{
		prevTx:=prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature=nil
		txCopy.Inputs[inId].Pubkey=prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID=txCopy.Hash()
		txCopy.Inputs[inId].Pubkey=nil

		r,s,err:=ecdsa.Sign(rand.Reader,&privKey,txCopy.ID)
		Handle(err)
		signature :=append(r.Bytes(),s.Bytes()...)
		tx.Inputs[inId].Signature=signature
	}
}

func (tx *Transaction)Verify(prevTXs map[string]Transaction) bool{
	if tx.IsCoinbase(){//coinbase needn't to be Signed
		return true
	}
	for _,in :=range tx.Inputs{
		if prevTXs[hex.EncodeToString(in.ID)].ID==nil{
			log.Panic("ERROR:Previous transaction is not correct")
		}
	}
	txCopy:=tx.TrimmedCopy()
	curve :=elliptic.P256()
	for inId,in:=range txCopy.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].Pubkey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].Pubkey = nil

		r:=big.Int{}
		s:=big.Int{}
		sigLen:=len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen/2)])
		s.SetBytes(in.Signature[(sigLen/2):])

		x:=big.Int{}
		y:=big.Int{}
		keyLen:=len(in.Pubkey)
		x.SetBytes(in.Signature[:(keyLen/2)])
		y.SetBytes(in.Signature[(keyLen/2):])

		rawPubkey:=ecdsa.PublicKey{curve,&x,&y}
		if ecdsa.Verify(&rawPubkey,txCopy.ID,&r,&s)==false{
			return false
		}
	}
	return true
}

func (tx Transaction)String() string {//How is it called?
	var lines []string
	lines =append(lines,fmt.Sprintf("--Transaction %x:",tx.ID))
	for i,input :=range tx.Inputs{
		lines =append(lines,fmt.Sprintf("Input %d:",i))
		lines=append(lines,fmt.Sprintf("TXID:	%x",input.ID))
		lines=append(lines,fmt.Sprintf("Out:	%d",input.Out))
		lines=append(lines,fmt.Sprintf("Signature:%X",input.Signature))
		lines=append(lines,fmt.Sprintf("pubKey:	%x",input.Pubkey))
	}
	for i,output:=range tx.Outputs {
		lines=append(lines,fmt.Sprintf("Output %d:",i))
		lines=append(lines,fmt.Sprintf("Value:	%d",output.Value))
		lines=append(lines,fmt.Sprintf("Script:	%x",output.PubKeyHash))
	}
	return strings.Join(lines,"\n")
}



