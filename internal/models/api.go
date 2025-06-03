package models

import (
	"blockchain-service/internal/blockchain"
	"encoding/hex"
) 

type BlockAPI struct{
	Hash string `json:"hash"`
	PrevHash string `json:"prevHash"`
	Nonce int `json:"nonce"`
	Timestamp int64 `json:"timestamp"`
	Data BlockDataAPI `json:"data"`
}

type BlockDataAPI struct{
	Hash string `json:"hash"`
	DocumentID string `json:"momId"`
	NotaryID string `json:"notaryId"`
	UserID string `json:"userId"`
	CNPJ string `json:"cnpj"`
}

func (bd *BlockDataAPI) ToBlockData() (*blockchain.BlockData, error){
	hashBytes, err := hex.DecodeString(bd.Hash)
	if err != nil{
		return nil, err
	}
	blockData := blockchain.BlockData{
		Hash: hashBytes,
		DocumentID: bd.DocumentID,
		NotaryID: bd.NotaryID,
		UserID: bd.UserID,
		CNPJ: bd.CNPJ,
	}

	return &blockData, nil
}

func FromBlockData(data *blockchain.BlockData) BlockDataAPI{
	hashString := hex.EncodeToString(data.Hash)
	blockAPI := BlockDataAPI{
		Hash: hashString,
		DocumentID: data.DocumentID,
		NotaryID: data.NotaryID,
		UserID: data.UserID,
		CNPJ: data.CNPJ,
	}

	return blockAPI
}

func FromBlock(block *blockchain.Block) BlockAPI{
	hashString := hex.EncodeToString(block.Hash)
	prevHashString := hex.EncodeToString(block.PrevHash)
	dataAPI := FromBlockData(&block.Data)

	blockAPI := BlockAPI{
		Hash: hashString,
		PrevHash: prevHashString,
		Nonce: block.Nonce,
		Timestamp: block.Timestamp,
		Data: dataAPI,
	}

	return blockAPI
}
