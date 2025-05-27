package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger/v4"
)

const(
	dbPath = "./tmp/blocks"
)

type BlockChain struct{
	LastHash []byte
	height uint64 
	Database *badger.DB
}


func (chain *BlockChain) CreateInsertBlock(data string) *Block{
	var lastHash []byte
	
	err := chain.Database.View(func(txn *badger.Txn) error{
		item, err := txn.Get([]byte("lh"))
		Handle(err)

		err = item.Value(func(val []byte) error{
			lastHash = val 
			return err
		})
		return err
	})
	Handle(err)

	block := CreateBlock(data, lastHash)
	chain.InsertBlock(block)
	return block
}

func (chain *BlockChain) InsertBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error{
		err := txn.Set([]byte(block.Hash), block.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), block.Hash)
		Handle(err)

		chain.LastHash = block.Hash
		return err 
	})
	Handle(err)
	chain.height += 1
}

func (chain *BlockChain) Height() uint64{
	return chain.height
}

func (chain *BlockChain) ContainsBlock(hash  []byte) bool{
	var block *Block
	iter := chain.Iterator()

	for{
		block = iter.Next()

		if bytes.Equal(block.Hash, hash){
			return true
		}
		if len(block.PrevHash) == 0{
			break
		}
	}
	return false
}

func (chain *BlockChain) GetBlockByHash(hash []byte) *Block{
	var block *Block 
	iter := chain.Iterator()

	for{
		block = iter.Next()

		if bytes.Equal(block.Hash, hash){
			return block
		}

		if len(block.Hash) == 0{
			break
		}
	}
	return nil
}

func (chain *BlockChain) ListBlocks() []*Block{
	blocks := []*Block{}
	var block *Block 

	iter := chain.Iterator()
	counter := 1
	for{
		block = iter.Next()
		blocks = append(blocks, block)
		log.Printf("\nBLOCK %d: %s", counter, hex.EncodeToString(block.Hash))
		log.Printf("PREVHASH: %s\n", hex.EncodeToString(block.PrevHash))
		counter++
		if len(block.PrevHash) == 0{
			break
		}
	}

	return blocks
}


func InitBlockChain() *BlockChain{
	var lastHash []byte 

	opts := badger.DefaultOptions(dbPath)
	opts.ValueLogFileSize = 1 << 27

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error{
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound{
			log.Println("No Existing blockchain found, creating one...")
			genesis := Genesis()

			err := txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash
			return err
		} 
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err = item.Value(func(val []byte) error{
			lastHash = val
			return err
		})

		return err
	})
		
	blockchain := BlockChain{lastHash, 1,db}
	return &blockchain
}


type BlockChainIterator struct{
	CurrentHash []byte 
	Database *badger.DB
}

func (chain *BlockChain) Iterator() *BlockChainIterator{
	iter := &BlockChainIterator{chain.LastHash, chain.Database}
	return iter
}

func (iter *BlockChainIterator) Next() *Block{
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error{
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		err = item.Value(func(val []byte) error{
			block = Deserialize(val)
			return err
		})
		return err
	})	
	Handle(err)
		
	iter.CurrentHash = block.PrevHash

	return block
}
