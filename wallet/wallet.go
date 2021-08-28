package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const(
	checksumLength =4
	version =byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}

func NewKeyPair()(ecdsa.PrivateKey,[]byte){//the aim of this function is also to generate a wallet
	curve:=elliptic.P256()

	private,err:=ecdsa.GenerateKey(curve,rand.Reader)
	if err!=nil{
		log.Panic(err)
	}
	pub:=append(private.PublicKey.X.Bytes(),private.PublicKey.Y.Bytes()...)
	return  *private,pub
}

func MakeWallet() *Wallet  {
	private,public:=NewKeyPair()
	wallet:=Wallet{private,public}
	return &wallet
}

func PublicKeyHash(pubKey []byte)[]byte{
	pubHash :=sha256.Sum256(pubKey)

	hasher:=ripemd160.New()
	_,err:=hasher.Write(pubHash[:])
	if err!=nil{
		log.Panic(err)
	}
	publicRipMD :=hasher.Sum(nil)
	return publicRipMD
}

func Checksum(payload []byte) []byte  {
	firstHash :=sha256.Sum256(payload)
	secondHash :=sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength]
}

func (w Wallet)Address() []byte {
	pubHash:=PublicKeyHash(w.PublicKey)

	versionedHash:=append([]byte{version},pubHash...)
	checksum:=Checksum(versionedHash)

	fullHash:=append(versionedHash,checksum...)
	address:=Base58Encode(fullHash)

	//fmt.Printf("pub key:%x\n",w.PublicKey)
	//fmt.Printf("pub hash:%x\n",pubHash)
	//fmt.Printf("address:%x\n",address)

	return address
}

//Address:14LErwM2aHhdsDym6PkyutyG9ZSm51UHXc
//FullHash:00|248bd9e7a51b7dd07aba9766a7c62d5020790280|2bc6c767
//[version] 00
//[Pub Key Hash]248bd9e7a51b7dd07aba9766a7c62d5020790280
//[CheckSum]2bc6c767

func ValidateAddress(address string)bool{
	FullHash:=Base58Decode([]byte(address))
	actualChecksum:=FullHash[len(FullHash)-checksumLength:]
	version:=FullHash[0]
	pubKeyHash:=FullHash[1:len(FullHash)-checksumLength]
	targetChecksum:=Checksum(append([]byte{version},pubKeyHash...))

	return bytes.Compare(actualChecksum,targetChecksum)==0
}

