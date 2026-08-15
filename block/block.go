package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"zxcoin/merkle"
	"zxcoin/transaction"
)

type BlockHeader struct {
	PrevHash  [32]byte
	Nonce     uint64
	RootHash  [32]byte
	Timestamp uint32
}

type Block struct {
	Header       BlockHeader
	Transactions []transaction.Transaction
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

func (b *Block) Mine() {
	for {
		hash := b.Header.Hash()

		valid := true

		for i := range b.Difficulty {
			if hash[i] != 0 {
				valid = false
				break
			}
		}

		if valid {
			return
		}

		b.Header.Nonce++
	}
}

func (b *Block) CalculateRootHash() {
	merkleTree := merkle.BuildMerkleTree(b.Transactions)
	b.Header.RootHash = merkleTree.Hash
}
