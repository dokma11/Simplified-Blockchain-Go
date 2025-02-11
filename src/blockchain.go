package main

type Blockchain struct {
	Blocks []*Block
}

func NewGenesisBlock() *Block {
	genesisBlock := NewBlock(nil, "Genesis Block")

	return genesisBlock
}

func NewBlockchain() *Blockchain {
	blockchain := Blockchain{}
	blockchain.Blocks = append(blockchain.Blocks, NewGenesisBlock())
	return &blockchain
}

func (bc *Blockchain) AddBlock(data string) {
	bc.Blocks = append(bc.Blocks, NewBlock(bc.Blocks[len(bc.Blocks)-1], data)) // previous block, data
}

func (bc *Blockchain) IsValid() bool {
	// Skip the genesis block since it stores no valuable data and has no previous block hash
	for i := 1; i < len(bc.Blocks); i++ {
		if bc.Blocks[i].Hash != bc.Blocks[i].CalculateHash() {
			return false
		}
		if bc.Blocks[i].PreviousHash != bc.Blocks[i-1].Hash {
			return false
		}
	}

	return true
}
