package odds

import (
	"math/big"
)

///////////
// MERGE //
///////////

/*
Merges all the listed odds objects into the main object. Returns a copy of the
base odds object to facilitated chaining.
*/
func (o *Odds[D, H]) Merge(objects ...*Odds[D, H]) *Odds[D, H] {
	for _, obj := range objects {

		if obj == nil {
			continue
		}

		for _, entry := range obj.Map {
			existingEntry := o.Map[entry.Hash]
			if existingEntry != nil {
				existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
			} else {
				o.Map[entry.Hash] = entry
			}
		}
		o.Total.Add(o.Total, obj.Total)
	}

	return o
}

/*
Merges all the listed odds objects into the main object. Returns a copy of the
base odds object to facilitated chaining.
*/
func (o *Odds[D, H]) Merge_Combine(objects ...*Odds[D, H]) *Odds[D, H] {
	for _, obj := range objects {

		if obj == nil {
			continue
		}

		for _, entry := range obj.Map {
			existingEntry := o.Map[entry.Hash]
			if existingEntry != nil {
				existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
				existingEntry.Data = o.CombineFunction(existingEntry.Data, entry.Data)
			} else {
				o.Map[entry.Hash] = entry
			}
		}
		o.Total.Add(o.Total, obj.Total)
	}

	return o
}

/*
Merges all the listed odds objects into the main object. Returns a copy of the
base odds object to facilitated chaining.
*/
func (o *Odds[D, H]) Merge_CombineInPlace(objects ...*Odds[D, H]) *Odds[D, H] {
	for _, obj := range objects {

		if obj == nil {
			continue
		}

		for _, entry := range obj.Map {
			existingEntry := o.Map[entry.Hash]
			if existingEntry != nil {
				existingEntry.Weight.Add(existingEntry.Weight, entry.Weight)
				o.CombineInPlaceFunction(existingEntry.Data, entry.Data)
			} else {
				o.Map[entry.Hash] = entry
			}
		}
		o.Total.Add(o.Total, obj.Total)
	}

	return o
}

//////////////
// CONVOLVE //
//////////////

/*
For every combination in "o" and each odds in "objects", apply
o.ConvolveFunction to get an array of new convolved entries to replace in "o"
*/
func (o *Odds[D, H]) Convolve(objects ...*Odds[D, H]) *Odds[D, H] {
	for _, entry := range o.Entries() {
		for _, objEntry := range objects[0].Entries() {

			/*
				Convolve the two entries data together resulting in a weight map
				representing the percentage of the combined weight each newEntry
				gets.
			*/
			newEntryArray := o.ConvolveFunction(o, entry, objEntry)

			/*
				Determine the total weight of objects returned from the
				convolve function
			*/
			total := big.NewInt(0)
			for _, newEntry := range newEntryArray {
				total.Add(total, newEntry.Weight)
			}

			// New weight that all the newEntries have to fit in
			newWeight := big.NewInt(0)
			newWeight.Mul(entry.Weight, objEntry.Weight)

			if total.Cmp(big.NewInt(1)) == 1 {
				o.Scale(total)
			}

			o.RemoveEntry(entry)
			for _, newEntry := range newEntryArray {
				o.Add(newEntry.Data, newEntry.Weight.Mul(newEntry.Weight, newWeight))
			}

		}
	}

	o.Reduce()

	if len(objects) == 1 {
		return o
	}

	return o.Convolve(objects[1:]...)
}

/*
 */
func (o *Odds[D, H]) ConvolveInPlace(objects ...*Odds[D, H]) *Odds[D, H] {

	for _, entry := range o.Map {
		for _, objEntry := range objects[0].Map {

			/*
				Convolve the two entries data together resulting in the entry
				in "o" being modified according to the convolve function.
			*/
			o.ConvolveInPlaceFunction(o, entry, objEntry)
		}
	}

	if len(objects) == 1 {
		return o
	}

	return o.Convolve(objects[1:]...)
}
