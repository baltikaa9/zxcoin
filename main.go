package main

import (
	// "crypto/ecdsa"
	"fmt"
	// "time"
)

func main() {
	// utxoDB := UTXODB{
	// 	UTXOKey{"123", 0}: TxOutput{5, "1"},
	// 	UTXOKey{"124", 0}: TxOutput{5, "2"},
	// }

	// block0 := Block{
	// 	BlockHeader{
	// 		"",
	// 		0,
	// 		"",
	// 		uint32(time.Now().Unix()),
	// 	},
	// 	make([]Transaction, 0),
	// }

	// input := TxInput{"123", 0, "123"}
	// input2 := TxInput{"124", 0, "123"}
	// output1 := TxOutput{5, "1"}
	// output2 := TxOutput{5, "2"}

	// t1 := NewTransaction([]TxInput{input}, []TxOutput{output1})
	// t2 := NewTransaction([]TxInput{input2}, []TxOutput{output2})

	// block1 := NewBlock(
	// 	block0,
	// 	10,
	// 	[]Transaction{t1, t2},
	// )

	// block2 := NewBlock(
	// 	block1,
	// 	10,
	// 	[]Transaction{t1, t2},
	// )

	// blockchain := Blockchain{}

	// err := blockchain.AddBlock(block0, utxoDB)
	// if err != nil {
	// 	fmt.Printf("%v", err)
	// 	return
	// }

	// err = blockchain.AddBlock(block1, utxoDB)
	// if err != nil {
	// 	fmt.Printf("%v", err)
	// 	return
	// }

	// // err = blockchain.AddBlock(block2, utxoDB)
	// // if err != nil {
	// // 	fmt.Printf("%v", err)
	// // 	return
	// // }

	// fmt.Printf("%v", blockchain)

	wallet := NewWallet()

	t0 := NewTransaction(
		nil,
		[]TxOutput{{5, wallet.PublicKey}},
	)

	t := NewTransaction(
		[]TxInput{{t0.Hash(), 0, Signature{}}},
		[]TxOutput{{5, wallet.PublicKey}},
	)

	tHash := t.Hash()

	for i := range t.Inputs {
		t.Inputs[i].Sign(wallet.PrivateKey, tHash)
	}

	fmt.Printf("%v\n", t.Inputs[0].Signature)

	oldOutput := t0.Outputs[t.Inputs[0].OutIndex]
	
	fmt.Printf("%v", t.Inputs[0].Verify(oldOutput.PublicKey, tHash))
}
