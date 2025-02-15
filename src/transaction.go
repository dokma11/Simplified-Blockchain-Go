package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

const subsidy = 10

type Transaction struct {
	ID   string
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = string(hash[:])
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TxID) == 0 && tx.Vin[0].Vout == -1
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txIn := TXInput{"", -1, data}
	txOut := TXOutput{subsidy, to}
	tx := Transaction{"", []TXInput{txIn}, []TXOutput{txOut}}
	tx.SetID()

	return &tx
}

func NewUTXOTransaction(sender, recipient string, amount int, bc *Blockchain) *Transaction {
	if amount <= 0 {
		log.Panic("ERROR: Transaction must have an amount greater than zero")
	}

	availableBalance, spendableOutputs := bc.FindSpendableOutputs(sender, amount)
	if availableBalance < amount {
		log.Panic("ERROR: Not enough funds")
	}

	inputs := createInputs(spendableOutputs, sender)
	outputs := createOutputs(sender, recipient, amount, availableBalance)

	tx := Transaction{"", inputs, outputs}
	tx.SetID()

	return &tx
}

func createInputs(spendableOutputs map[string][]int, sender string) []TXInput {
	var inputs []TXInput
	for txID, outputIndexes := range spendableOutputs {
		for _, index := range outputIndexes {
			inputs = append(inputs, TXInput{txID, index, sender})
		}
	}
	return inputs
}

func createOutputs(sender, recipient string, amount, availableBalance int) []TXOutput {
	outputs := []TXOutput{{amount, recipient}}

	change := availableBalance - amount
	if change > 0 {
		outputs = append(outputs, TXOutput{change, sender})
	}

	return outputs
}
