package blockchain_logic

import (
	"bytes"
	"encoding/gob"

	"github.com/AlonzaT/blockchain/utils"
	"github.com/AlonzaT/blockchain/wallet"
)

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

func (outs TxOutputs) Serialize() []byte {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(outs)
	utils.HandleErr(err)

	return buffer.Bytes()
}

func (outs TxOutputs) DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&outputs)
	utils.HandleErr(err)

	return outputs
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

//Check to see if the signature is equal to the data (account owner)
/*func (in *TxInput) CanUnlock(data string) bool {
	//Check to see if the signature is equal to the data
	return in.Signature == data
}*/

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {

	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
