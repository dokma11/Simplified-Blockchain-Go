package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type Block struct {
	Index        int
	Data         string
	Timestamp    string
	Hash         string
	PreviousHash string
}

// CalculateHash TODO: mozda dodati neku validaciju pa da se vrati error ako treba
func (b *Block) CalculateHash() string {
	hashData := strconv.Itoa(b.Index) + ":" + b.Timestamp + ":" + b.Data + ":" + b.PreviousHash
	hash := sha256.Sum256([]byte(hashData))
	return hex.EncodeToString(hash[:])
}

func NewBlock(previousBlock *Block, data string) *Block {
	block := Block{
		Timestamp: time.Now().String(),
		Data:      data,
	}

	if previousBlock != nil {
		block.Index = previousBlock.Index + 1
		block.PreviousHash = previousBlock.Hash
	} else {
		block.PreviousHash = "-"
	}

	block.Hash = block.CalculateHash()

	return &block
}
