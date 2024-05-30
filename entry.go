package odds

import (
	"math/big"
	"sort"
)

//////////////////////
// ENTRY DEFINITION //
//////////////////////

type Entry[D any, H comparable] struct {
	Hash   H
	Data   D
	Weight *big.Int
}

/////////////////////////////
// INSTANTIATION FUNCTIONS //
/////////////////////////////

// Create a new entry with the specified weight and data
func (o *Odds[D, H]) NewEntry(data D, weight *big.Int) *Entry[D, H] {
	return &Entry[D, H]{o.HashFunction(data), data, weight}
}

// Create a new entry with the specified weight and data
func (o *Odds[D, H]) NewEntryWithHash(hash H, data D, weight *big.Int) *Entry[D, H] {
	return &Entry[D, H]{hash, data, weight}
}

/*
Create a new entry with the specified weight and data. Creates a copy of both
the data and the weight before creating
*/
func (o *Odds[D, H]) CopyEntry(entry *Entry[D, H]) *Entry[D, H] {
	return &Entry[D, H]{
		o.HashFunction(entry.Data),
		o.CopyFunction(entry.Data),
		new(big.Int).Set(entry.Weight),
	}
}

////////////////////////
// ODDS ENTRY METHODS //
////////////////////////

/*
Check the the data exists in "o" using the the hash function of "o". Returns nil
if there is no entry found.
*/
func (o *Odds[D, H]) Exists(data D) *Entry[D, H] {
	return o.Map[o.HashFunction(data)]
}

/*
Get a list of all the entries, sorted by the Less function
*/
func (o *Odds[D, H]) Entries() []*Entry[D, H] {
	entries := []*Entry[D, H]{}
	for _, entry := range o.Map {
		entries = append(entries, entry)
	}
	return entries
}

/*
Get a list of all the entries, sorted by the their contribution
*/
func (o *Odds[D, H]) EntriesByWeight() []*Entry[D, H] {
	entries := []*Entry[D, H]{}
	for _, entry := range o.Map {
		entries = append(entries, entry)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Weight.Cmp(entries[j].Weight) < 0
	})

	return entries
}
