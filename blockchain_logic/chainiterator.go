package blockchain_logic

import (
	"github.com/AlonzaT/blockchain/utils"
	"github.com/dgraph-io/badger"
)

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

//Convert Blockchain struct to Blockchain Iterator struct
func (chain *BlockChain) Iterator() *BlockChainIterator {
	i := &BlockChainIterator{chain.LastHash, chain.Database}

	return i
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	var encodedBlock []byte

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		utils.HandleErr(err)
		valErr := item.Value(func(val []byte) error {
			encodedBlock = val
			return nil
		})
		block = Deserialize(encodedBlock)

		return valErr
	})
	utils.HandleErr(err)

	iter.CurrentHash = block.PrevHash

	return block
}

func (i *BlockChain) Flush() {
	i.Database.DropAll()
}
