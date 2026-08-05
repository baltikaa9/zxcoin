package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type BlockHeader struct {
	PrevHash  string
	Nonce     int
	RootHash  string
	Timestamp uint32
}

type Block struct {
	Header       BlockHeader
	Transactions []Transaction
}

func NewBlock(prevBlock Block, nonce int, transactions []Transaction) Block {
	prevData := prevBlock.Header.PrevHash + strconv.Itoa(prevBlock.Header.Nonce) + prevBlock.Header.RootHash + strconv.Itoa(int(prevBlock.Header.Timestamp))
	prevHash := sha256.Sum256([]byte(prevData))

	header := BlockHeader{
		hex.EncodeToString(prevHash[:]),
		nonce,
		getRootHash(transactions),
		uint32(time.Now().Unix()),
	}

	return Block{header, transactions}
}

func getRootHash(transactions []Transaction) string {
	return ""
}
