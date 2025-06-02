package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math"
	"math/big"
)

const Dificulty = 12

type ProofOfWork struct{
	Block *Block 
	Target *big.Int
}

func (pow *ProofOfWork) Run() (int, []byte){
	var intHash big.Int
	var hash [32]byte 

	nonce := 0
	for nonce < math.MaxInt64{
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		log.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])
		
		if intHash.Cmp(pow.Target) == -1{
				break 
		}
		nonce++
	}

	return nonce, hash[:]
}

func NewProof(b *Block) *ProofOfWork{
	target := big.NewInt(1)
	target.Lsh(target, uint(256 - Dificulty))

	pow := &ProofOfWork{b, target}
	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte{
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.Data.Serialize(),
			ToHex(int64(nonce)),
			ToHex(int64(Dificulty)),
		},
		[]byte{},
	)

	return data
}

func ToHex(num int64) []byte{
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil{
		log.Panic(err)
	}
	return buff.Bytes()
}

func (pow *ProofOfWork) Validate() bool{
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}
