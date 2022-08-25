package blockchain_logic

// Take data from the block

// Create a counter (nounce) which starts at 0

// Create a hash of the data plus counter

// Check the hash to see if it meets a set of requirements

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	"github.com/AlonzaT/blockchain/utils"
)

// In a real blockchain the difficulty would
// increase over time
const Difficulty = 12

//ProofOfWork has a pointer to a block and a pointer to a target
type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

// Takes a block and pairs it with a Target
func NewProof(b *Block) *ProofOfWork {
	//create a new Target as a big interger
	target := big.NewInt(1)

	//Subtract the difficulty from 256
	//which is the number of bytes inside one of our hashes
	//and use the target to shift the bytes over
	//LSH function is short hand for left shift
	target.Lsh(target, uint(256-Difficulty))

	//Add block and target to proof of work
	pow := &ProofOfWork{b, target}

	//return the proof of work
	//has a block and target stored in it
	return pow
}

//Turns data into a block of bytes
func (pow *ProofOfWork) InitData(nounce int) []byte {
	//Joins all the bites together in different structs
	data := bytes.Join(
		[][]byte{
			/*It's joining each of the following together
			to create a new byte*/

			//Starting with the Block prevHash
			pow.Block.PrevHash,

			//Transaction hashes from the Block
			pow.Block.HashTransactions(),

			//then the nounce as bytes
			ToHex(int64(nounce)),

			//then the difficulty as bytes
			ToHex(int64(Difficulty)),
		},
		//Puts all the bytes into one slice of bytes
		[]byte{},
	)

	//All the bytes together creates the data
	return data
}

//returns a nounce and a hash
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nounce := 0

	for nounce < math.MaxInt64 {
		//Prepares the data
		data := pow.InitData(nounce)

		//Sum256 hashes the data.
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)

		//SetBytes interprets buf as the bytes of a
		//big-endian unsigned integer,
		//sets z to that value, and returns z.
		intHash.SetBytes(hash[:])

		//The loop is running until the statement comes back
		// as true
		// if int hash is less than the proof of work break out
		// the loop because we signed the block
		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			//else increase the nounce
			nounce++
		}
	}
	fmt.Println()
	
	//return the nounce and the hash
	return nounce, hash[:]
}

//validates the blocks
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nounce)

	//Sum256 returns the SHA256 checksum of the data.
	hash := sha256.Sum256(data)

	//SetBytes interprets buf as the bytes of a
	//big-endian unsigned integer,
	//sets z to that value, and returns z.
	intHash.SetBytes(hash[:])

	//compares inthash to pow.target
	//if inthash is less than pow.target return true
	return intHash.Cmp(pow.Target) == -1
}

//
func ToHex(num int64) []byte {
	//creates a new bytes buffer
	buff := new(bytes.Buffer)

	//Writes to the bytes buffer in big Endian format
	err := binary.Write(buff, binary.BigEndian, num)
	utils.HandleErr(err)

	//return the bytes
	return buff.Bytes()
}
