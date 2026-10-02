// Package block определяет структуру блока и его заголовка, а также proof-of-work майнинг и вычисление корня дерева Меркла.
package block

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/baltikaa9/zxcoin/transaction"
	"github.com/baltikaa9/zxcoin/types"
)

type BlockHeader struct {
	PrevHash  types.Hash
	Nonce     uint64
	RootHash  types.Hash
	Timestamp uint32
}

type Block struct {
	Header       BlockHeader
	Transactions []transaction.Transaction
}

func (bh BlockHeader) serialize() []byte {
	buf := bh.PrevHash[:]
	buf = append(buf, bh.RootHash[:]...)
	buf = binary.BigEndian.AppendUint64(buf, bh.Nonce)
	buf = binary.BigEndian.AppendUint32(buf, bh.Timestamp)

	return buf
}

func (bh BlockHeader) Hash() types.Hash {
	return sha256.Sum256(bh.serialize())
}

func (b *Block) Mine(difficulty uint64) {
	for {
		hash := b.Header.Hash()

		valid := true

		for i := range difficulty {
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

func (b Block) ValidateProofOfWork(difficulty uint64) error {
	hash := b.Header.Hash()

	for i := range difficulty {
		if hash[i] != 0 {
			return InvalidNonceError{BlockHash: hash, Difficulty: difficulty}
		}
	}

	return nil
}

func (b *Block) CalculateRootHash() error {
	txHashes, err := b.getTxHashes()

	if err != nil {
		return err
	}

	b.Header.RootHash = buildMerkleTree(txHashes).Hash

	return nil
}

func (b Block) ValidateRootHash() error {
	txHashes, err := b.getTxHashes()

	if err != nil {
		return err
	}

	if b.Header.RootHash != buildMerkleTree(txHashes).Hash {
		return InvalidMerkleRootError{}
	}

	return nil
}

func (b Block) getTxHashes() ([]types.Hash, error) {
	txHashes := make([]types.Hash, 0, len(b.Transactions))

	for _, tx := range b.Transactions {
		hash, err := tx.Hash()

		if err != nil {
			return nil, err
		}

		txHashes = append(txHashes, hash)
	}

	return txHashes, nil
}
