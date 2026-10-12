package wallet

import "github.com/baltikaa9/zxcoin/core/domain/utxo"

type ReservationTracker struct {
	reserved map[utxo.ID]struct{}
}

func NewReservationTracker() *ReservationTracker {
	return &ReservationTracker{reserved: make(map[utxo.ID]struct{})}
}

func (t *ReservationTracker) IsReserved(id utxo.ID) bool {
	_, exists := t.reserved[id]
	return exists
}

func (t *ReservationTracker) Reserve(id utxo.ID) {
	t.reserved[id] = struct{}{}
}

func (t *ReservationTracker) Release(id utxo.ID) {
	delete(t.reserved, id)
}
