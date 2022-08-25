package blockchain_logic

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

//The Block Representation
type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nounce       int
	Height       int
	Timestamp    int64
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}

	tree := NewMerkleTree(txHashes)

	return tree.RootNode.Data
}

//Create Block
func CreateBlock(txs []*Transaction, prevHash []byte, height int) *Block {
	//initiate block
	block := &Block{[]byte{}, txs, prevHash, 0, height, time.Now().Unix()}
	//Cretaes a new proof of work with the block
	pow := NewProof(block)
	nounce, hash := pow.Run()

	//Add block Hash to block
	block.Hash = hash[:]
	//Add the block nounce
	block.Nounce = nounce

	//retturn the block
	return block
}

//Create the Genesis Block
func GenesisBlock(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0)
}

//Turns whole block into bytes
func (b *Block) Serialize() []byte {
	var res bytes.Buffer

	//Encode data a gob of gytes
	encoder := gob.NewEncoder(&res)

	//Encodes block to bytes
	err := encoder.Encode(b)
	HandleErr(err)

	return res.Bytes()

}

//Turns block back into intelegible data
func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)
	HandleErr(err)

	return &block
}

func HandleErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}
