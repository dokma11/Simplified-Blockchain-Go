package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
)

func main() {
	blockchain := NewBlockchain()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("Enter command (add <data> / validate / save <filepath> / exit):")
		scanner.Scan()
		input := scanner.Text()
		command := strings.Fields(input)

		if len(command) == 0 {
			continue
		}

		switch command[0] {
		case "add":
			if len(command) < 2 {
				fmt.Println("Usage: add <data>")
				continue
			}

			blockData := strings.Join(command[1:], " ")
			blockchain.AddBlock(blockData)

			fmt.Println("Block added.")
		case "validate":
			if blockchain.IsValid() {
				fmt.Println("Blockchain is valid.")
			} else {
				fmt.Println("Blockchain is invalid.")
			}
		case "save":
			if len(command) < 2 {
				fmt.Println("Usage: save <filepath>")
				continue
			}

			filepath := fp.Join(command[1:]...)

			if fp.Ext(filepath) != ".json" {
				filepath += ".json"
			}

			blockchainData, err := json.MarshalIndent(blockchain, "", "\t")
			if err != nil {
				panic(err)
			}

			if err := os.WriteFile(filepath, blockchainData, 0666); err != nil {
				panic(err)
			}

			fmt.Printf("Blockchain saved [%s].\n", filepath)
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("Unknown command")
		}
	}
}
