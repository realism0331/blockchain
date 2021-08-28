package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile="./tmp/wallets_%s.data"

type Wallets struct {
	Wallets map[string] *Wallet// address as key
}

func(ws *Wallets)LoadFile(nodeId string) error{
	walletFile:=fmt.Sprintf(walletFile,nodeId)
	if _,err:=os.Stat(walletFile);os.IsNotExist(err){
		return err
	}

	var wallets Wallets

	fileContent,err:=ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	gob.Register(elliptic.P256())
	decoder:=gob.NewDecoder(bytes.NewReader(fileContent))
	err =decoder.Decode(&wallets)//to decode the fileContent into wallets
	if err != nil {
		log.Panic(err)
	}
	ws.Wallets=wallets.Wallets

	return nil
}

func (ws *Wallets)SaveFile(nodeId string)  {
	var content bytes.Buffer
	walletFile:=fmt.Sprintf(walletFile,nodeId)

	gob.Register(elliptic.P256())//this encoder is based on elliptic
	encoder:=gob.NewEncoder(&content)//Append a encoder to the "content" cache
	err:= encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err =ioutil.WriteFile(walletFile,content.Bytes(),0644)//The third parameter is used to set permissions
	if err != nil {
		log.Panic(err)
	}
}

func CreateWallets(nodeId string)(*Wallets,error){
	wallets:=Wallets{}
	wallets.Wallets=make(map[string]*Wallet)//Slices need to be created separately

	err:=wallets.LoadFile(nodeId)

	return &wallets,err
}

func (ws *Wallets)GetWallet(address string)Wallet  {
	return *ws.Wallets[address]
}

func(ws *Wallets) GetAllAddresses()[] string  {
	var addresses []string
	for address:=range ws.Wallets {//If there is only one parameter, it would accept the key of map;if  two,first will accept the key,second will accept the value
		addresses=append(addresses,address)
	}
	return addresses
}

func(ws *Wallets)AddWallet()string{
	wallet:=MakeWallet()
	address:=fmt.Sprintf("%s",wallet.Address())

	ws.Wallets[address]=wallet
	return address
}