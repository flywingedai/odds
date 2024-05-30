package odds

import (
	"fmt"
	"math/big"
	"sync"
)

/////////////////////
// ODDS DEFINITION //
/////////////////////

type Odds[D any, H comparable] struct {

	// Map stores all the entries for the odds
	Map map[H]*Entry[D, H]

	// Total weight of entries in the odds map
	Total *big.Int

	// CUSTOMIZABLE FUNCTIONS //

	// How each entry.Data is hashed
	HashFunction func(D) H

	// How to copy each entry.Data
	CopyFunction func(D) D

	// How two entry.Data are added together to produce a third
	CombineFunction func(D, D) D

	// How two entry.Data are added together such that the first argument
	// is modified by the second
	CombineInPlaceFunction func(D, D)

	// How two entry.Data are convolved together to produce a third
	ConvolveFunction func(*Odds[D, H], *Entry[D, H], *Entry[D, H]) []*Entry[D, H]

	// How two entry.Data are convolved together such that the first argument
	// is modified by the second
	ConvolveInPlaceFunction func(*Odds[D, H], *Entry[D, H], *Entry[D, H])

	// How each data object should be displayed
	DisplayFunction func(D) string

	// Used for sync operations
	lock sync.Mutex
}

// Creates a full copy of the odds object
func (o *Odds[D, H]) Copy() *Odds[D, H] {
	newOdds := NewOddsFromReference(o)

	for _, entry := range o.Entries() {
		newEntry := o.CopyEntry(entry)
		newOdds.Map[entry.Hash] = newEntry
		newOdds.Total.Add(newOdds.Total, newEntry.Weight)
	}

	return newOdds
}

////////////////////////
// OPTIONS DEFINTIONS //
////////////////////////

/*
Options for creating a new Odds Object
*/
type OddsOptions[D any, H comparable] struct {
	HashFunction            func(D) H
	CopyFunction            func(D) D
	CombineFunction         func(D, D) D
	CombineInPlaceFunction  func(D, D)
	ConvolveFunction        func(*Odds[D, H], *Entry[D, H], *Entry[D, H]) []*Entry[D, H]
	ConvolveInPlaceFunction func(*Odds[D, H], *Entry[D, H], *Entry[D, H])
	DisplayFunction         func(D) string
}

// OPTIONS CONSTRUCTORS //

/*
Create a new OddsOptions with a specified hash function
*/
func NewOptions[D any, H comparable](hashFunction func(D) H) *OddsOptions[D, H] {
	return &OddsOptions[D, H]{
		HashFunction:    hashFunction,
		DisplayFunction: func(d D) string { return fmt.Sprint(d) },
	}
}

/*
Specify the hash function in the options
*/
func (options *OddsOptions[D, H]) WithHash(hashFunction func(D) H) *OddsOptions[D, H] {
	options.HashFunction = hashFunction
	return options
}

/*
Specify the copy function in the options
*/
func (options *OddsOptions[D, H]) WithCopy(copyFunction func(D) D) *OddsOptions[D, H] {
	options.CopyFunction = copyFunction
	return options
}

/*
Specify the add function in the options
*/
func (options *OddsOptions[D, H]) WithAdd(combineFunction func(D, D) D) *OddsOptions[D, H] {
	options.CombineFunction = combineFunction
	return options
}

/*
Specify the addInPlace function in the options
*/
func (options *OddsOptions[D, H]) WithAddInPlace(combineInPlaceFunction func(D, D)) *OddsOptions[D, H] {
	options.CombineInPlaceFunction = combineInPlaceFunction
	return options
}

/*
Specify the convolve function in the options
*/
func (options *OddsOptions[D, H]) WithConvolve(
	convolveFunction func(*Odds[D, H], *Entry[D, H], *Entry[D, H]) []*Entry[D, H],
) *OddsOptions[D, H] {
	options.ConvolveFunction = convolveFunction
	return options
}

/*
Specify the convolveInPlace function in the options
*/
func (options *OddsOptions[D, H]) WithConvolveInPlace(
	convolveInPlaceFunction func(*Odds[D, H], *Entry[D, H], *Entry[D, H]),
) *OddsOptions[D, H] {
	options.ConvolveInPlaceFunction = convolveInPlaceFunction
	return options
}

/*
Specify the display function in the options
*/
func (options *OddsOptions[D, H]) WithDisplay(displayFunction func(D) string) *OddsOptions[D, H] {
	options.DisplayFunction = displayFunction
	return options
}

/////////////////////////////
// INSTANTIATION FUNCTIONS //
/////////////////////////////

/*
Create a new Odds object with arbitrary weight precision based on the options
provided. If not all functions are provided, some features will cause crashes
if used.
*/
func (options *OddsOptions[D, H]) Odds() *Odds[D, H] {
	return &Odds[D, H]{
		Map:   map[H]*Entry[D, H]{},
		Total: big.NewInt(0),

		HashFunction:            options.HashFunction,
		CopyFunction:            options.CopyFunction,
		CombineFunction:         options.CombineFunction,
		CombineInPlaceFunction:  options.CombineInPlaceFunction,
		ConvolveFunction:        options.ConvolveFunction,
		ConvolveInPlaceFunction: options.ConvolveInPlaceFunction,
		DisplayFunction:         options.DisplayFunction,

		lock: sync.Mutex{},
	}
}

/*
Create a new Odds object based on the reference.
*/
func NewOddsFromReference[D any, H comparable](reference *Odds[D, H]) *Odds[D, H] {
	newOdds := NewOptions(reference.HashFunction).Odds()

	newOdds.CopyFunction = reference.CopyFunction
	newOdds.CombineFunction = reference.CombineFunction
	newOdds.CombineInPlaceFunction = reference.CombineInPlaceFunction
	newOdds.ConvolveFunction = reference.ConvolveFunction
	newOdds.ConvolveInPlaceFunction = reference.ConvolveInPlaceFunction
	newOdds.DisplayFunction = reference.DisplayFunction

	return newOdds
}

/*
Create a new Odds object based using "o" as the reference.
*/
func (o *Odds[D, H]) AsReference() *Odds[D, H] {
	return NewOddsFromReference(o)
}

///////////////////
// MODIFICATIONS //
///////////////////

/*
Specify the copy function in the odds
*/
func (o *Odds[D, H]) WithHash(hashFunction func(D) H) *Odds[D, H] {
	o.HashFunction = hashFunction
	o.UpdateHashes()
	return o
}

/*
Updates all the hashes of the object in place. Returns the odds object to
facilitate chaining. Is useful to run after modifying data objects individually.
*/
func (o *Odds[D, H]) UpdateHashes() *Odds[D, H] {
	newEntries := map[H]*Entry[D, H]{}

	for _, entry := range o.Map {
		newHash := o.HashFunction(entry.Data)
		entry.Hash = newHash
		existingEntry, exists := newEntries[newHash]
		if exists {
			existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
		} else {
			newEntries[newHash] = entry
		}
	}

	o.Map = newEntries
	return o
}

/*
Specify the copy function in the odds
*/
func (o *Odds[D, H]) WithCopy(copyFunction func(D) D) *Odds[D, H] {
	o.CopyFunction = copyFunction
	return o
}

/*
Specify the add function in the odds
*/
func (o *Odds[D, H]) WithAdd(combineFunction func(D, D) D) *Odds[D, H] {
	o.CombineFunction = combineFunction
	return o
}

/*
Specify the addInPlace function in the odds
*/
func (o *Odds[D, H]) WithAddInPlace(combineInPlaceFunction func(D, D)) *Odds[D, H] {
	o.CombineInPlaceFunction = combineInPlaceFunction
	return o
}

/*
Specify the convolve function in the odds
*/
func (o *Odds[D, H]) WithConvolve(
	convolveFunction func(*Odds[D, H], *Entry[D, H], *Entry[D, H]) []*Entry[D, H],
) *Odds[D, H] {
	o.ConvolveFunction = convolveFunction
	return o
}

/*
Specify the convolveInPlace function in the odds
*/
func (o *Odds[D, H]) WithConvolveInPlace(
	convolveInPlaceFunction func(*Odds[D, H], *Entry[D, H], *Entry[D, H]),
) *Odds[D, H] {
	o.ConvolveInPlaceFunction = convolveInPlaceFunction
	return o
}

/*
Specify the display function in the odds
*/
func (o *Odds[D, H]) WithDisplay(displayFunction func(D) string) *Odds[D, H] {
	o.DisplayFunction = displayFunction
	return o
}
