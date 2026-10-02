package wallet

import "github.com/baltikaa9/zxcoin/core/domain/utxo"

type ReservationTracker struct {
	reserved map[utxo.UTXOID]struct{}
}

func NewReservationTracker() *ReservationTracker {
	return &ReservationTracker{reserved: make(map[utxo.UTXOID]struct{})}
}

func (t *ReservationTracker) IsReserved(id utxo.UTXOID) bool {
	_, exists := t.reserved[id]
	return exists
}

func (t *ReservationTracker) Reserve(id utxo.UTXOID) {
	t.reserved[id] = struct{}{}
}

func (t *ReservationTracker) Release(id utxo.UTXOID) {
	delete(t.reserved, id)
}
