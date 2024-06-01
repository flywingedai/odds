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
	addFlags OddsFlags,
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

	mergeFunction := o.Merge
	if addFlags&Add_Combine > 0 {
		mergeFunction = o.Merge_Combine
	} else if addFlags&Add_CombineInPlace > 0 {
		mergeFunction = o.Merge_CombineInPlace
	}

	for i, extendedOdds := range oddsArray {
		scaleFactor := new(big.Int).Mul(multipledExtendedWeight, entryArray[i].Weight)
		scaleFactor.Div(scaleFactor, extendedOdds.Total)
		extendedOdds.Scale(scaleFactor)
		mergeFunction(extendedOdds)
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
	addFlags OddsFlags,
) *Odds[D, H] {

	type workerDataStruct struct {
		subTotal    *big.Int
		entryWeight *big.Int
		scaleFactor chan *big.Int
	}

	workerDataChan := make(chan *workerDataStruct, workers)
	completedOdds := make(chan *Odds[D, H], workers)

	// Define the main worker function that processes each entry
	entries := o.Entries()
	workerFunction := func(startIndex, endIndex int) {

		/*
			Reimplementation of o.ExtendOdds with a bunch of extra parameters
			which are useful for parallelization. TODO: Update Odds.ExtendOdds
			to be easier to use here
		*/
		multipledExtendedWeight := big.NewInt(1)
		oddsArray := []*Odds[D, H]{}

		totalEntryWeight := big.NewInt(0)
		newOdds := NewOddsFromReference(o)

		for i := startIndex; i < endIndex; i++ {
			entry := entries[i]
			extendedOdds := extendFunction(newOdds, entry).Reduce()
			oddsArray = append(oddsArray, extendedOdds)
			multipledExtendedWeight.Mul(multipledExtendedWeight, extendedOdds.Total)
			totalEntryWeight.Add(totalEntryWeight, entry.Weight)
		}

		for i, extendedOdds := range oddsArray {
			scaleFactor := new(big.Int).Mul(multipledExtendedWeight, entries[startIndex+i].Weight)
			scaleFactor.Div(scaleFactor, extendedOdds.Total)
			extendedOdds.Scale(scaleFactor)

			if addFlags&Add_Combine > 0 {
				newOdds.Merge_Combine(extendedOdds)
			} else if addFlags&Add_CombineInPlace > 0 {
				newOdds.Merge_CombineInPlace(extendedOdds)
			} else {
				newOdds.Merge(extendedOdds)
			}
		}

		newOdds.Reduce().UpdateHashes()

		// Inform the main process about the size of the resulting odds
		data := &workerDataStruct{newOdds.Total, totalEntryWeight, make(chan *big.Int)}
		workerDataChan <- data

		/*
			Once the worker is done, we want to send back on the completed
			channel the computed new odds, correctly scaled based on the total
			computed weight
		*/
		completedOdds <- newOdds.Scale(<-data.scaleFactor)

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
		go workerFunction(currentStartIndex, endIndex)
		currentStartIndex = endIndex
	}

	/*
		Determine the total weight of the resulting maps and send the value to
		each individual worker.
	*/
	totalWeight := big.NewInt(1)
	workerData := []*workerDataStruct{}
	for i := 0; i < workers; i++ {
		data := <-workerDataChan
		workerData = append(workerData, data)
		totalWeight.Mul(totalWeight, data.subTotal)
	}

	// Calculate all the scale factors
	scaleFactors := []*big.Int{}
	var gcd *big.Int
	for _, d := range workerData {
		scaleFactor := new(big.Int).Quo(new(big.Int).Mul(d.entryWeight, totalWeight), d.subTotal)

		if gcd == nil {
			gcd = new(big.Int).Set(scaleFactor)
		} else {
			gcd.GCD(nil, nil, gcd, scaleFactor)
		}

		scaleFactors = append(scaleFactors, scaleFactor)
	}

	for i, d := range workerData {
		adjustedFactor := new(big.Int).Quo(scaleFactors[i], gcd)
		d.scaleFactor <- adjustedFactor
	}

	// Fetch the completed odds and combine into o
	o.Clear()

	mergeFunction := o.Merge
	if addFlags&Add_Combine > 0 {
		mergeFunction = o.Merge_Combine
	} else if addFlags&Add_CombineInPlace > 0 {
		mergeFunction = o.Merge_CombineInPlace
	}

	for i := 0; i < workers; i++ {
		completed := <-completedOdds
		mergeFunction(completed)
	}

	return o.Reduce_Parallel(workers)
}
