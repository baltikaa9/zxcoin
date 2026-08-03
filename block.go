package main

type BlockHeader struct {
	PrevHash string
	Nonce int
	RootHash string
	Timestamp uint32
}

type Block struct {
	Header BlockHeader
	Transactions []Transaction
}