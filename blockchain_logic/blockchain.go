package blockchain_logic

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlonzaT/blockchain/utils"
	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "First Transaction from Genesis"
)

//Chain of Blocks
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

func DBExist(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}

func DeserializeTransaction(data []byte) Transaction {
	var tx Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	utils.HandleErr(err)
	return tx
}

func ContinueBlockChain(nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBExist(path) == false {
		fmt.Println("No exsisting blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(path)
	opts.Dir = path
	opts.ValueDir = path
	opts.Logger = nil

	db, err := openDB(path, opts)
	HandleErr(err)

	dbErr := db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		HandleErr(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})

		return err
	})
	HandleErr(dbErr)

	chain := BlockChain{lastHash, db}
	return &chain
}

//Initialize blockchain
func InitBlockChain(address, nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBExist(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	var lastHash []byte

	opts := badger.DefaultOptions(path)
	opts.Dir = path
	opts.ValueDir = path
	opts.Logger = nil

	db, err := openDB(path, opts)
	HandleErr(err)

	dbErr := db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := GenesisBlock(cbtx)
		fmt.Println("Genesis Created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		HandleErr(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})
	HandleErr(dbErr)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

//Add block to the chain
func (chain *BlockChain) MineBlock(TXs []*Transaction) *Block {
	var lastHash []byte
	var lastBlockData []byte
	var lastHeight int

	for _, tx := range TXs {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Invalid Transaction")
		}
	}

	//Get last Hash So we can create a new block
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.HandleErr(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})

		item, err = txn.Get(lastHash)
		err = item.Value(func(val []byte) error {
			lastBlockData = val
			return nil
		})
		utils.HandleErr(err)

		lastBlock := Deserialize(lastBlockData)
		lastHeight = lastBlock.Height

		return err
	})
	HandleErr(err)

	newBlock := CreateBlock(TXs, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		HandleErr(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	HandleErr(err)

	return newBlock
}

func (chain *BlockChain) AddBlock(block *Block) {
	var lastHash []byte
	var lastBlockData []byte

	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		utils.HandleErr(err)

		item, err := txn.Get([]byte("lh"))
		utils.HandleErr(err)
		iErr := item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.HandleErr(iErr)

		item, err = txn.Get(lastHash)
		utils.HandleErr(err)
		iErr2 := item.Value(func(val []byte) error {
			lastBlockData = val
			return nil
		})
		utils.HandleErr(iErr2)

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			utils.HandleErr(err)
			chain.LastHash = block.Hash
		}

		return nil
	})
	utils.HandleErr(err)
}

func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block
	var blockData []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			err := item.Value(func(val []byte) error {
				blockData = val
				return nil
			})
			utils.HandleErr(err)

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *BlockChain) GetBestHeight() int {
	var lastBlock Block
	var lastHash []byte
	var lastBlockData []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.HandleErr(err)
		valErr := item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.HandleErr(valErr)

		item, err = txn.Get(lastHash)
		utils.HandleErr(err)
		val2Err := item.Value(func(val []byte) error {
			lastBlockData = val
			return nil
		})
		utils.HandleErr(val2Err)

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	utils.HandleErr(err)

	return lastBlock.Height
}

func (chain *BlockChain) FindUTXO( /*pubKeyHash []byte*/ ) map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	//unspentTransactions := chain.FindUnspentTransactions(pubKeyHash)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}

	}

	return UTXO
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		utils.HandleErr(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		utils.HandleErr(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
