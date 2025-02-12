package main

import (
	"fmt"
	"github.com/boltdb/bolt"
)

func main() {
	bc := NewBlockchain()
	defer func(DB *bolt.DB) {
		err := DB.Close()
		if err != nil {
			fmt.Printf("Error occurred while closing the DB: %s", err)
		}
	}(bc.DB)

	cli := CLI{bc}
	cli.Run()
}
