package blockchain

import(
	"log"
	"github.com/dgraph-io/badger/v4"
)

const(
	dbPath = "./tmp/blocks"
)

type BlockChain struct{
	LastHash []byte	
	Database *badger.DB
}

func (chain *BlockChain) AddBlock(data string){
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

	err = chain.Database.Update(func(txn *badger.Txn) error{
		err := txn.Set([]byte(block.Hash), block.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), block.Hash)
		Handle(err)

		chain.LastHash = block.Hash
		return err 
	})
	Handle(err)
}


func InitBlockChain() *BlockChain{
	var lastHash []byte 

	opts := badger.DefaultOptions(dbPath)

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
		
	blockchain := BlockChain{lastHash, db}
	return &blockchain
}
