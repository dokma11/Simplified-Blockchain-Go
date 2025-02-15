package main

import "bytes"

type TXInput struct {
	TxID      string
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	return bytes.Compare(HashPubKey(in.PubKey), pubKeyHash) == 0
}
