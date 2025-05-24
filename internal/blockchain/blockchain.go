package blockchain

import (
	"encoding/hex"
	"log"
	"os"
	"runtime"

	"github.com/dgraph-io/badger/v4"
)

const(
	dbPath = "./tmp/blocks"
	dbFile = "./tmp/blocks/MANIFEST"
	genesisData = "Genesis Data"
)

type BlockChain struct{
	LastHash []byte	
	Database *badger.DB
}


func (chain *BlockChain) AddBlock(transactions []*Transaction){
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

	block := CreateBlock(transactions, lastHash)

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

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction{
	var unspentTxs []Transaction
	spentTxs := make(map[string][]int)
	
	iter := chain.Iterator()
	for{
		block := iter.Next()
		
		for _, tx := range block.Transactions{
			txID := hex.EncodeToString(tx.ID)

			Outputs:
			for outIdx, out := range tx.Outputs{
				if spentTxs[txID] != nil{
					for _, spentOut := range spentTxs[txID]{
						if spentOut != outIdx{
							continue Outputs
						}
					}
				}

				if out.CanBeUnlocked(address){
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinbase() == false{
				for _, in := range tx.Inputs{
					if in.CanUnlock(address){
						inTxID := hex.EncodeToString(in.ID)
						spentTxs[inTxID] = append(spentTxs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0{
			break
		}
	}

	return unspentTxs
}

func (chain *BlockChain) FindUTXO(address string) []TxOutput{
	var UTXOs []TxOutput
	unspentTxs := chain.FindUnspentTransactions(address)
	
	for _, tx := range unspentTxs{
		for _, out := range tx.Outputs{
			if out.CanBeUnlocked(address){
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int){
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)
	acc := 0
	
	Work:
	for _, tx := range unspentTxs{
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs{
			if out.CanBeUnlocked(address) && acc < amount{
				acc += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if acc >= amount{
					break Work
				}
			}
		}
	}

	return acc, unspentOuts
}

func InitBlockChain(address string) *BlockChain{
	var lastHash []byte 
	
	if DBExists(){
		log.Println("BlockChain Already Exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error{
		log.Println("No Existing blockchain found, creating one...")
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)

		err := txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash
		return err
	})
		
	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func ContinueBlockChain(address string) *BlockChain{
	if DBExists() == false{
		log.Println("No existing blockchain")
		runtime.Goexit()
	}

	var lastHash []byte 

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error{
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		
		err = item.Value(func(val []byte) error{
			lastHash = val 
			return err
		})

		return err
	})
	Handle(err)
		
	blockchain := BlockChain{lastHash, db}
	return &blockchain

}

func DBExists() bool{
	if _, err := os.Stat(dbFile); os.IsNotExist(err){
		return false 
	}
	return true
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
		
	iter.CurrentHash = block.Hash

	return block
}
