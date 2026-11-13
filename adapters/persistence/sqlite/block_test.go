package sqlite

import (
	"errors"
	"testing"

	"github.com/baltikaa9/zxcoin/adapters/persistence"
	"github.com/baltikaa9/zxcoin/core/domain/block"
	"github.com/baltikaa9/zxcoin/core/testutil"

	_ "modernc.org/sqlite"
)

func TestSave(t *testing.T) {
	db, err := testutil.OpenTestSQLiteDB("../../../migrations/*.up.sql")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewBlock(db)

	_, height, exists, err := repo.Tip()
	if err != nil {
		t.Fatalf("get tip: %v", err)
	}
	if !exists {
		t.Fatal("genesis is not exists")
	}

	newBlock := block.Block{}
	if err := repo.Save(newBlock, height+1); err != nil {
		t.Fatalf("save: %v", err)
	}

	b, newHeight, exists, err := repo.Tip()
	if err != nil {
		t.Fatalf("get tip: %v", err)
	}
	if !exists {
		t.Fatal("genesis is not exists")
	}

	if b.Header.Hash() != newBlock.Header.Hash() || newHeight != height+1 {
		t.Fatalf(
			"invalid new block, got: hash=%s, height=%d; want: hash=%s, height=%d",
			b.Header.Hash().String(),
			newHeight,
			newBlock.Header.Hash().String(),
			height+1,
		)
	}
}

func TestSave_InvalidHeight(t *testing.T) {
	db, err := testutil.OpenTestSQLiteDB("../../../migrations/*.up.sql")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewBlock(db)

	firstBlock, height, exists, err := repo.Tip()
	if err != nil {
		t.Fatalf("get tip: %v", err)
	}
	if !exists {
		t.Fatal("genesis is not exists")
	}

	newBlock := block.Block{}

	err = repo.Save(newBlock, 10)

	testutil.CheckError[persistence.InvalidHeightError](t, err)

	typedError, _ := errors.AsType[persistence.InvalidHeightError](err)

	if typedError.Got != 10 || typedError.Want != height+1 {
		t.Fatal(err)
	}

	secondBlock, newHeight, exists, err := repo.Tip()
	if err != nil {
		t.Fatalf("get tip: %v", err)
	}
	if !exists {
		t.Fatal("genesis is not exists")
	}

	if secondBlock.Header.Hash() != firstBlock.Header.Hash() || newHeight != height {
		t.Fatal("tip changed")
	}
}

func TestSave_RepeatedHeight(t *testing.T) {
	db, err := testutil.OpenTestSQLiteDB("../../../migrations/*.up.sql")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := NewBlock(db)

	firstBlock, height, exists, err := repo.Tip()
	if err != nil {
		t.Fatalf("get tip: %v", err)
	}
	if !exists {
		t.Fatal("genesis is not exists")
	}

	newBlock := block.Block{}

	err = repo.Save(newBlock, height)

	testutil.CheckError[persistence.InvalidHeightError](t, err)

	typedError, _ := errors.AsType[persistence.InvalidHeightError](err)

	if typedError.Got != height || typedError.Want != height+1 {
		t.Fatal(err)
	}

	secondBlock, newHeight, exists, err := repo.Tip()
	if err != nil {
		t.Fatalf("get tip: %v", err)
	}
	if !exists {
		t.Fatal("genesis is not exists")
	}

	if secondBlock.Header.Hash() != firstBlock.Header.Hash() || newHeight != height {
		t.Fatal("tip changed")
	}
}
