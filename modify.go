package odds

import (
	"math/big"
)

// Types for merging, combining, and convolving
type ModifyFlag int

const (
	Modify_Default ModifyFlag = 1 << iota
	Modify_Combine
	Modify_CombineInPlace
	Modify_ConvolveInPlace
)

///////////////
// ADDITIONS //
///////////////

/*
Add new data with a specified weight to the odds. This will not copy the
data being passed in.
*/
func (o *Odds[D, H]) Add(data D, weight *big.Int) {
	entry := o.NewEntry(data, weight)
	existingEntry := o.Map[entry.Hash]
	if existingEntry != nil {
		existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
	} else {
		o.Map[entry.Hash] = entry
	}
	o.Total.Add(o.Total, entry.Weight)
}

/*
Add new entry to the odds. This will not copy the data being passed in.
*/
func (o *Odds[D, H]) AddEntry(entry *Entry[D, H]) {
	existingEntry := o.Map[entry.Hash]
	if existingEntry != nil {
		existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
	} else {
		o.Map[entry.Hash] = entry
	}
	o.Total.Add(o.Total, entry.Weight)
}

/*
Add new data with a specified weight to the odds. This will not copy the
data being passed in.
*/
func (o *Odds[D, H]) Add_Combine(data D, weight *big.Int) {
	entry := o.NewEntry(data, weight)
	existingEntry := o.Map[entry.Hash]
	if existingEntry != nil {
		existingEntry.Weight.Add(existingEntry.Weight, weight)
		existingEntry.Data = o.CombineFunction(existingEntry.Data, entry.Data)
	} else {
		o.Map[entry.Hash] = entry
	}
	o.Total.Add(o.Total, weight)
}

/*
Add new entry to the odds. This will not copy the data being passed in.
*/
func (o *Odds[D, H]) AddEntry_Combine(entry *Entry[D, H]) {
	existingEntry := o.Map[entry.Hash]
	if existingEntry != nil {
		existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
		existingEntry.Data = o.CombineFunction(existingEntry.Data, entry.Data)
	} else {
		o.Map[entry.Hash] = entry
	}
	o.Total.Add(o.Total, entry.Weight)
}

/*
Add new data with a specified weight to the odds. This will not copy the
data being passed in.
*/
func (o *Odds[D, H]) Add_CombineInPlace(data D, weight *big.Int) {
	entry := o.NewEntry(data, weight)
	existingEntry := o.Map[entry.Hash]
	if existingEntry != nil {
		existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
		o.CombineInPlaceFunction(existingEntry.Data, entry.Data)
	} else {
		o.Map[entry.Hash] = entry
	}
	o.Total.Add(o.Total, entry.Weight)
}

/*
Add new entry to the odds. This will not copy the data being passed in.
*/
func (o *Odds[D, H]) AddEntry_CombineInPlace(entry *Entry[D, H]) {
	existingEntry := o.Map[entry.Hash]
	if existingEntry != nil {
		existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
		o.CombineInPlaceFunction(existingEntry.Data, entry.Data)
	} else {
		o.Map[entry.Hash] = entry
	}
	o.Total.Add(o.Total, entry.Weight)
}

/*
Add new odds "newOdds" to "o". newOdds has a corresponding weight which is the
proportion of weight the whole of newOdds is intended to take up in "o" relative
to the rest of "o". Both "o" and newOdds will be scaled accordingly to preserve
this intended relationship
*/
func (o *Odds[D, H]) AddOdds(newOdds *Odds[D, H], weight *big.Int) *Odds[D, H] {
	if weight.Cmp(big.NewInt(0)) == 0 {
		panic("")
	}

	gcd := new(big.Int).GCD(nil, nil, newOdds.Total, weight)
	reducedTotal := new(big.Int).Div(newOdds.Total, gcd)
	reducedWeight := new(big.Int).Div(weight, gcd)

	o.Scale(reducedTotal)
	scaledNewOdds := newOdds.Scale(reducedWeight)
	for _, entry := range scaledNewOdds.Map {
		o.AddEntry(entry)
	}
	return o
}

/////////////
// REMOVAL //
/////////////

/*
Removes the data associated with the given hash from "o". Returns the weight
removed from "o". The *big.Int removed is the actual weight, from the removed
entry, so be aware of any mutations of that value.
*/
func (o *Odds[D, H]) RemoveHash(hash H) *big.Int {
	existingEntry := o.Map[hash]
	if existingEntry == nil {
		return nil
	}

	delete(o.Map, hash)
	o.Total.Sub(o.Total, existingEntry.Weight)
	return existingEntry.Weight
}

/*
Removes the data from "o". Returns the weight removed from "o". The *big.Int
removed is the actual weight, from the removed entry, so be aware of any
mutations of that value.
*/
func (o *Odds[D, H]) RemoveData(data D) *big.Int {
	return o.RemoveHash(o.HashFunction(data))
}

/*
Removes the entry from "o". Returns the weight removed from "o". The *big.Int
removed is the actual weight, from the removed entry, so be aware of any
mutations of that value.
*/
func (o *Odds[D, H]) RemoveEntry(entry *Entry[D, H]) *big.Int {
	return o.RemoveHash(entry.Hash)
}

/*
Tries to remove a whole subset from "o". If only partial removal occurs, the
total amount of weight removed will be returned.
*/
func (o *Odds[D, H]) RemoveSubset(subset *Odds[D, H]) *big.Int {
	originalTotal := big.NewInt(0).Set(o.Total)

	// We try to remove each entry from the subset from "o"
	for hash, entry := range subset.Map {
		existingEntry := o.Map[hash]
		if existingEntry == nil {
			continue
		}

		o.Total.Sub(o.Total, o.RemoveEntry(entry))
	}

	// This is how we track how much weight in "o" has been removed. It is
	return originalTotal.Sub(originalTotal, o.Total)
}

//////////////////
// REPLACEMENTS //
//////////////////

/*
Given a subset that exists in "o", remove all entries from that subset (or
corresponding weight) and replace it with an entire new odds object.
*/
func (o *Odds[D, H]) ReplaceSubsetWithOdds(subset *Odds[D, H], newOdds *Odds[D, H]) *Odds[D, H] {
	amountRemoved := o.RemoveSubset(subset)
	if amountRemoved.Cmp(big.NewInt(0)) == 0 {
		panic("")
	}
	o.AddOdds(newOdds, amountRemoved)
	return o
}

/*
Given a subset that exists in "o", remove all entries from that subset (or
corresponding weight) and replace it with a single entry that represents the
entire weight of the subset removed.
*/
func (o *Odds[D, H]) ReplaceSubsetWithData(subset *Odds[D, H], data D) *Odds[D, H] {
	amountRemoved := o.RemoveSubset(subset)
	if amountRemoved.Cmp(big.NewInt(0)) == 0 {
		panic("")
	}
	o.Add(data, amountRemoved)
	return o
}

/*
Given a hash that exists in "o", remove the corresponding entry from "o" and
replace it with an entire new odds object.
*/
func (o *Odds[D, H]) ReplaceHashWithOdds(hash H, newOdds *Odds[D, H]) *Odds[D, H] {
	amountRemoved := o.RemoveHash(hash)
	if amountRemoved.Cmp(big.NewInt(0)) == 0 {
		panic("")
	}
	o.AddOdds(newOdds, amountRemoved)
	return o
}

/*
Given a hash that exists in "o", remove the corresponding entry from "o" and
replace it with an entire new odds object.
*/
func (o *Odds[D, H]) ReplaceHashWithData(hash H, data D) *Odds[D, H] {
	amountRemoved := o.RemoveHash(hash)
	if amountRemoved.Cmp(big.NewInt(0)) == 0 {
		panic("")
	}
	o.Add(data, amountRemoved)
	return o
}

/*
Given data that exists in "o", remove the corresponding entry from "o" and
replace it with an entire new odds object.
*/
func (o *Odds[D, H]) ReplaceDataWithOdds(dataToRemove D, newOdds *Odds[D, H]) *Odds[D, H] {
	return o.ReplaceHashWithOdds(o.HashFunction(dataToRemove), newOdds)
}

/*
Given data that exists in "o", remove the corresponding entry from "o" and
replace it with an entire new odds object.
*/
func (o *Odds[D, H]) ReplaceDataWithData(remove, data D) *Odds[D, H] {
	return o.ReplaceHashWithData(o.HashFunction(data), data)
}

/*
Given and entry that exists in "o", replace it with an entire new odds object.
Scales everything appropriately so no precision is lost.
*/
func (o *Odds[D, H]) ReplaceEntryWithOdds(entry *Entry[D, H], newOdds *Odds[D, H]) *Odds[D, H] {
	amountRemoved := o.RemoveEntry(entry)
	if amountRemoved.Cmp(big.NewInt(0)) == 0 {
		panic("")
	}
	o.AddOdds(newOdds, amountRemoved)
	return o
}

/*
Given and entry that exists in "o", replace it with an entire new odds object.
Scales everything appropriately so no precision is lost.
*/
func (o *Odds[D, H]) ReplaceEntryWithData(entry *Entry[D, H], data D) *Odds[D, H] {
	amountRemoved := o.RemoveEntry(entry)
	if amountRemoved.Cmp(big.NewInt(0)) == 0 {
		panic("")
	}
	o.Add(data, amountRemoved)
	return o
}

/////////////////
// ADJUSTMENTS //
/////////////////

// Scales all the existing weights on the odds object by the given factor.
func (o *Odds[D, H]) Scale(factor *big.Int) *Odds[D, H] {
	for _, entry := range o.Map {
		entry.Weight.Mul(entry.Weight, factor)
	}
	o.Total.Mul(o.Total, factor)
	return o
}

/*
Finds the Greatest Common Divisor (GCD) off all the weights on the odds object,
and then divides each of the weights and the odds total by that divisor.
*/
func (o *Odds[D, H]) Reduce() *Odds[D, H] {
	if o.Total.Cmp(big.NewInt(0)) == 0 {
		return o
	}

	// Finds the gcd
	var gcd *big.Int
	for _, entry := range o.Map {
		if gcd == nil {
			gcd = big.NewInt(0).Set(entry.Weight)
		} else {
			gcd.GCD(nil, nil, gcd, entry.Weight)
		}
	}

	// Modify the weights and the total before returning
	for _, entry := range o.Map {
		entry.Weight.Div(entry.Weight, gcd)
	}
	o.Total.Div(o.Total, gcd)

	return o

}

/*
Finds the Greatest Common Divisor (GCD) off all the weights on the odds object,
and then divides each of the weights and the odds total by that divisor.
*/
func (o *Odds[D, H]) Reduce_Parallel(workers int) *Odds[D, H] {
	if o.Total.Cmp(big.NewInt(0)) == 0 {
		return o
	}

	weightsChannel := make(chan *big.Int, len(o.Map))
	workerGCDs := make(chan *big.Int)
	totalGCD := make(chan *big.Int)
	doneChannel := make(chan bool)

	entries := o.Entries()
	gcdWorker := func(startIndex, endIndex int) {

		var gcd *big.Int
		weights := []*big.Int{}

		for i := startIndex; i < endIndex; i++ {
			weight := entries[i].Weight
			weights = append(weights, weight)
			if gcd == nil {
				gcd = new(big.Int).Set(weight)
			} else {
				gcd.GCD(nil, nil, gcd, weight)
			}
		}

		workerGCDs <- gcd
		gcd = <-totalGCD

		for _, weight := range weights {
			weight.Quo(weight, gcd)
		}

		doneChannel <- true

	}

	processPerWorker := 1 + (len(entries) / workers)
	currentStartIndex := 0
	for i := 0; i < workers; i++ {
		if currentStartIndex == len(entries) {
			workers = i
			break
		}

		endIndex := currentStartIndex + processPerWorker
		if endIndex > len(entries) {
			endIndex = len(entries)
		}
		go gcdWorker(currentStartIndex, endIndex)
		currentStartIndex = endIndex
	}

	for _, entry := range o.Map {
		weightsChannel <- entry.Weight
	}
	close(weightsChannel)

	var gcd *big.Int
	for i := 0; i < workers; i++ {
		workerGCD := <-workerGCDs
		if gcd == nil {
			gcd = workerGCD
		} else {
			gcd.GCD(nil, nil, gcd, workerGCD)
		}
	}

	for i := 0; i < workers; i++ {
		totalGCD <- gcd
	}

	o.Total.Div(o.Total, gcd)
	for i := 0; i < workers; i++ {
		<-doneChannel
	}

	return o

}
