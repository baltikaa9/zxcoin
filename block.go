package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"time"
)

type BlockHeader struct {
	PrevHash  [32]byte
	Nonce     uint64
	RootHash  [32]byte
	Timestamp uint32
}

type Block struct {
	Header       BlockHeader
	Transactions []Transaction
	Difficulty   int
}

func (bh BlockHeader) Hash() [32]byte {
	var buf bytes.Buffer

	buf.Write(bh.PrevHash[:])
	buf.Write(bh.RootHash[:])

	binary.Write(&buf, binary.BigEndian, bh.Nonce)
	binary.Write(&buf, binary.BigEndian, bh.Timestamp)

	return sha256.Sum256(buf.Bytes())
}

func NewBlock(prevBlock Block, transactions []Transaction) Block {
	return Block{
		BlockHeader{
			prevBlock.Header.Hash(),
			0,
			getRootHash(transactions),
			uint32(time.Now().Unix()),
		},
		transactions,
		0,
	}
}

func getRootHash(transactions []Transaction) [32]byte {
	return sha256.Sum256([]byte{1, 2, 3})
}
