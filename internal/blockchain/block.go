package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct{
	Hash []byte  `json:"hash"`
	PrevHash []byte `json:"prev_hash"`
	Nonce int `json:"nonce"`
	Timestamp int64 `json:"timestamp"` 
	Data BlockData `json:"data"`
}

type BlockData struct{
	Hash []byte	`json:"hash"`
	DocumentID string	`json:"documentId"`
	NotaryID string `json:"notaryId"`
	UserID string	`json:"userId"`
	CNPJ string `json:"cnpj"`
}

func (bd *BlockData) Serialize() []byte{
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(bd)
	Handle(err)
	return res.Bytes()
}

func CreateBlock(data *BlockData, PrevHash []byte) *Block{
	block := &Block{
		[]byte{}, 
		PrevHash, 
		0,
		time.Now().UnixMilli(),
		*data, 
	}
	
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash 
	block.Nonce = nonce

	return block
}

func Genesis() *Block{
	blockData := BlockData{
		[]byte{},
		"Genesis",
		"Genesis",
		"Genesis",
		"Genesis",
	}
	return CreateBlock(&blockData, []byte{})
}

func (b *Block) Serialize() []byte{
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)
	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block{
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)
	Handle(err)

	return &block
}


func Handle(err error){
	if err != nil{
		log.Panic(err)
	}
}
