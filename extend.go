package odds

import (
	"math/big"
)

/*
For each entry in o, replace the data with the result of the specified extend
function on that entry.
*/
func (o *Odds[D, H]) Extend(
	extendFunction func(*Entry[D, H]) D,
) *Odds[D, H] {
	for _, entry := range o.Entries() {
		entry.Data = extendFunction(entry)
		entry.Hash = o.HashFunction(entry.Data)
	}
	return o.UpdateHashes()
}

/*
For each entry in o, get a new group of odds which should replace it based on
the provided extendFunction.
*/
func (o *Odds[D, H]) ExtendOdds(
	extendFunction func(*Odds[D, H], *Entry[D, H]) *Odds[D, H],
	modifyFlags ModifyFlag,
) *Odds[D, H] {

	multipledExtendedWeight := big.NewInt(1)
	oddsArray := []*Odds[D, H]{}
	entryArray := []*Entry[D, H]{}

	for _, entry := range o.Entries() {
		extendedOdds := extendFunction(o, entry).Reduce()
		oddsArray = append(oddsArray, extendedOdds)
		entryArray = append(entryArray, entry)
		multipledExtendedWeight.Mul(multipledExtendedWeight, extendedOdds.Total)
	}

	o.Map = map[H]*Entry[D, H]{}
	o.Total.Set(big.NewInt(0))
	for i, extendedOdds := range oddsArray {
		scaleFactor := new(big.Int).Mul(multipledExtendedWeight, entryArray[i].Weight)
		scaleFactor.Div(scaleFactor, extendedOdds.Total)
		extendedOdds.Scale(scaleFactor)

		if modifyFlags&Modify_Combine > 0 {
			o.Merge_Combine(extendedOdds)
		} else if modifyFlags&Modify_CombineInPlace > 0 {
			o.Merge_CombineInPlace(extendedOdds)
		} else {
			o.Merge(extendedOdds)
		}
	}

	return o.Reduce()
}

/*
Perform o.Extend in parallel based on the number of workers provided.
*/
func (o *Odds[D, H]) Extend_Parallel(
	extendFunction func(*Entry[D, H]) D,
	workers int,
) *Odds[D, H] {

	// Define the main worker function that processes each entry
	entryQueue := make(chan *Entry[D, H])
	done := make(chan bool)
	workerFunction := func() {
		for {
			entry, isOpen := <-entryQueue
			if isOpen {
				entry.Data = extendFunction(entry)
				entry.Hash = o.HashFunction(entry.Data)
			} else {
				break
			}
		}
		done <- true
	}

	// Parallel start all the workers
	for i := 0; i < workers; i++ {
		go workerFunction()
	}

	// Send the workers all the entries to process
	for _, entry := range o.Entries() {
		entryQueue <- entry
	}

	// Close the queue when all the workers are done and return the extended "o"
	close(entryQueue)
	for i := 0; i < workers; i++ {
		<-done
	}

	return o
}

/*
Perform o.ExtendOdds in parallel based on the number of workers provided.

mergeType = ["", "Combine", "In Place"]
*/
func (o *Odds[D, H]) ExtendOdds_Parallel(
	extendFunction func(*Odds[D, H], *Entry[D, H]) *Odds[D, H],
	workers int,
	modifyFlags ModifyFlag,
) *Odds[D, H] {

	entryQueue := make(chan *Entry[D, H], len(o.Map))
	newOddsWeights := make(chan *big.Int, workers)
	totalWeights := make(chan *big.Int, workers)
	completedOdds := make(chan *Odds[D, H], workers)

	// Define the main worker function that processes each entry
	workerFunction := func() {

		// IMPLEMENTATION 1

		multipledExtendedWeight := big.NewInt(1)
		oddsArray := []*Odds[D, H]{}
		entryArray := []*Entry[D, H]{}

		totalEntryWeight := big.NewInt(0)
		newOdds := NewOddsFromReference(o)

		for {
			entry, isOpen := <-entryQueue
			if isOpen {
				extendedOdds := extendFunction(newOdds, entry).Reduce()
				oddsArray = append(oddsArray, extendedOdds)
				entryArray = append(entryArray, entry)
				multipledExtendedWeight.Mul(multipledExtendedWeight, extendedOdds.Total)
				totalEntryWeight.Add(totalEntryWeight, entry.Weight)
			} else {
				break
			}
		}

		recievedData := len(entryArray) > 0

		for i, extendedOdds := range oddsArray {
			scaleFactor := new(big.Int).Mul(multipledExtendedWeight, entryArray[i].Weight)
			scaleFactor.Div(scaleFactor, extendedOdds.Total)
			extendedOdds.Scale(scaleFactor)

			if modifyFlags&Modify_Combine > 0 {
				newOdds.Merge_Combine(extendedOdds)
			} else if modifyFlags&Modify_CombineInPlace > 0 {
				newOdds.Merge_CombineInPlace(extendedOdds)
			} else {
				newOdds.Merge(extendedOdds)
			}
		}

		newOdds.Reduce()

		// Inform the main process about the size of the resulting odds
		if recievedData {
			newOddsWeights <- newOdds.Total
		} else {
			newOddsWeights <- nil
		}

		/*
			Once the worker is done, we want to send back on the completed
			channel the computed new odds, correctly scaled based on the total
			computed weight
		*/
		totalWeight := <-totalWeights

		if recievedData {
			scaleFactor := new(big.Int).Div(totalEntryWeight.Mul(totalEntryWeight, totalWeight), newOdds.Total)
			completedOdds <- newOdds.Scale(scaleFactor).UpdateHashes()
		} else {
			completedOdds <- nil
		}

	}

	// Parallel start all the workers
	for i := 0; i < workers; i++ {
		go workerFunction()
	}

	/*
		Send the workers all the entries to process and close the entryQueue
		once all entries have been sent.
	*/
	for _, entry := range o.Entries() {
		entryQueue <- entry
	}
	close(entryQueue)

	/*
		Determine the total weight of the resulting maps and send the value to
		each individual worker.
	*/
	totalWeight := big.NewInt(1)
	for i := 0; i < workers; i++ {
		subTotal := <-newOddsWeights
		if subTotal == nil {
			continue
		}
		totalWeight.Mul(totalWeight, subTotal)
	}
	for i := 0; i < workers; i++ {
		totalWeights <- totalWeight
	}

	// Fetch the completed odds and combine into o
	o.Map = map[H]*Entry[D, H]{}
	o.Total.Set(big.NewInt(0))
	for i := 0; i < workers; i++ {
		completed := <-completedOdds
		if completed == nil {
			continue
		}

		if modifyFlags&Modify_Combine > 0 {
			o.Merge_Combine(completed)
		} else if modifyFlags&Modify_CombineInPlace > 0 {
			o.Merge_CombineInPlace(completed)
		} else {
			o.Merge(completed)
		}
	}

	return o.Reduce_Parallel(workers)
}
