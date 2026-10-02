package block

import (
	"crypto/sha256"
	"testing"

	"github.com/baltikaa9/zxcoin/core/domain/coin"
	"github.com/baltikaa9/zxcoin/core/domain/transaction"
	"github.com/baltikaa9/zxcoin/core/domain/types"
	"github.com/baltikaa9/zxcoin/core/testutil"
)

func TestBuildMerkleTree_OneNode(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)

	tx := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 5, Owner: publicKey}},
	}

	hash, err := tx.Hash()

	if err != nil {
		t.Fatal(err)
	}

	root := buildMerkleTree([]types.Hash{hash})

	if root.Hash != hash {
		t.Fatalf("root не совпадает с хешем единственной транзакции. Ожидалось: %v, получено: %v", hash, root.Hash)
	}
}

func TestBuildMerkleTree_TwoNodes(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)

	t1 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 5, Owner: publicKey}},
	}

	t2 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 10, Owner: publicKey}},
	}

	t1Hash, err := t1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	t2Hash, err := t2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	root := buildMerkleTree([]types.Hash{t1Hash, t2Hash})

	expectedHash := sha256.Sum256(append(t1Hash[:], t2Hash[:]...))

	if root.Hash != expectedHash {
		t.Fatalf("root не совпадает с хешем единственной транзакции. Ожидалось: %v, получено: %v", expectedHash, root.Hash)
	}
}

func TestBuildMerkleTree_ThreeNodes(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)

	t1 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 5, Owner: publicKey}},
	}

	t2 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 10, Owner: publicKey}},
	}

	t3 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 15, Owner: publicKey}},
	}

	t1Hash, err := t1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	t2Hash, err := t2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	t3Hash, err := t3.Hash()

	if err != nil {
		t.Fatal(err)
	}

	root := buildMerkleTree([]types.Hash{t1Hash, t2Hash, t3Hash})

	t12Hash := sha256.Sum256(append(t1Hash[:], t2Hash[:]...))
	t33Hash := sha256.Sum256(append(t3Hash[:], t3Hash[:]...))
	expectedHash := sha256.Sum256(append(t12Hash[:], t33Hash[:]...))

	if root.Hash != expectedHash {
		t.Fatalf("root не совпадает с хешем единственной транзакции. Ожидалось: %v, получено: %v", expectedHash, root.Hash)
	}
}

func TestBuildMerkleTree_Twice(t *testing.T) {
	_, publicKey := testutil.GenerateKeyPair(t)

	t1 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 5, Owner: publicKey}},
	}

	t2 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 10, Owner: publicKey}},
	}

	t3 := transaction.Transaction{
		Outputs: []coin.TxOutput{{Amount: 15, Owner: publicKey}},
	}

	t1Hash, err := t1.Hash()

	if err != nil {
		t.Fatal(err)
	}

	t2Hash, err := t2.Hash()

	if err != nil {
		t.Fatal(err)
	}

	t3Hash, err := t3.Hash()

	if err != nil {
		t.Fatal(err)
	}

	root1 := buildMerkleTree([]types.Hash{t1Hash, t2Hash, t3Hash})
	root2 := buildMerkleTree([]types.Hash{t1Hash, t2Hash, t3Hash})

	if root1.Hash != root2.Hash {
		t.Fatalf("root отличается для одинаковых транзакций. Ожидалось: %v, получено: %v", root1.Hash, root2.Hash)
	}
}
