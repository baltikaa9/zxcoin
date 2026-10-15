package dto

type UTXO struct {
	TxID     string `json:"tx_id"`
	OutIndex uint64 `json:"out_index"`
	Amount   uint64 `json:"amount"`
}

type UTXOResponse struct {
	UTXOs []UTXO `json:"utxos"`
	Sum   uint64 `json:"sum"`
}
