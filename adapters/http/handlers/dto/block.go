package dto

type Header struct {
	PrevHash  string `json:"prev_hash"`
	Nonce     uint64 `json:"nonce"`
	RootHash  string `json:"root_hash"`
	Timestamp uint32 `json:"timestamp"`
}

type Block struct {
	Header       Header        `json:"header"`
	Transactions []Transaction `json:"transactions"`
}
