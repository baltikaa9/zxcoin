package merkle

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"zxcoin/coin"
	"zxcoin/transaction"
)

func TestBuildMerkleTree_OneNode(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	tx := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}

	root := BuildMerkleTree([]transaction.Transaction{tx})

	if root.Hash != tx.Hash() {
		t.Fatalf("root не совпадает с хешем единственной транзакции. Ожидалось: %v, получено: %v", tx.Hash(), root.Hash)
	}
}

func TestBuildMerkleTree_TwoNodes(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	t1 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}

	t2 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 10, PublicKey: publicKey}},
	}

	root := BuildMerkleTree([]transaction.Transaction{t1, t2})
	t1Hash := t1.Hash()
	t2Hash := t2.Hash()
	expectedHash := sha256.Sum256(append(t1Hash[:], t2Hash[:]...))

	if root.Hash != expectedHash {
		t.Fatalf("root не совпадает с хешем единственной транзакции. Ожидалось: %v, получено: %v", expectedHash, root.Hash)
	}
}

func TestBuildMerkleTree_ThreeNodes(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	t1 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}

	t2 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 10, PublicKey: publicKey}},
	}

	t3 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 15, PublicKey: publicKey}},
	}

	root := BuildMerkleTree([]transaction.Transaction{t1, t2, t3})
	t1Hash := t1.Hash()
	t2Hash := t2.Hash()
	t3Hash := t3.Hash()
	t12Hash := sha256.Sum256(append(t1Hash[:], t2Hash[:]...))
	t33Hash := sha256.Sum256(append(t3Hash[:], t3Hash[:]...))
	expectedHash := sha256.Sum256(append(t12Hash[:], t33Hash[:]...))

	if root.Hash != expectedHash {
		t.Fatalf("root не совпадает с хешем единственной транзакции. Ожидалось: %v, получено: %v", expectedHash, root.Hash)
	}
}

func TestBuildMerkleTree_Twice(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey

	t1 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 5, PublicKey: publicKey}},
	}

	t2 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 10, PublicKey: publicKey}},
	}

	t3 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 15, PublicKey: publicKey}},
	}

	root1 := BuildMerkleTree([]transaction.Transaction{t1, t2, t3})
	root2 := BuildMerkleTree([]transaction.Transaction{t1, t2, t3})

	if root1.Hash != root2.Hash {
		t.Fatalf("root отличается для одинаковых транзакций. Ожидалось: %v, получено: %v", root1.Hash, root2.Hash)
	}
}
