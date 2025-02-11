package main

type Blockchain struct {
	Blocks []*Block
}

func NewGenesisBlock() *Block {
	return NewBlock(nil, "Genesis Block")
}

func NewBlockchain() *Blockchain {
	blockchain := Blockchain{}
	blockchain.Blocks = append(blockchain.Blocks, NewGenesisBlock())
	return &blockchain
}

func (bc *Blockchain) AddBlock(data string) {
	bc.Blocks = append(bc.Blocks, NewBlock(bc.Blocks[len(bc.Blocks)-1], data)) // previous block, data
}
