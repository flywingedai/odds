package odds

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

type Hashable[T any] interface {
	Hash() string
	Combine(T) T
}

type Odds[D Hashable[D]] struct {
	Data    map[string]D
	Weights map[string]*big.Int
	Total   *big.Int
}

/////////////
// METHODS //
/////////////

/*
Add new data to the odds. If the data's "Hash()" already exists in the
odds, then the new weight will be added to the existing weight.
*/
func (o *Odds[D]) Add(data D, weight *big.Int) {
	hash := data.Hash()
	_, exists := o.Data[hash]
	if exists {
		existingWeights := o.Weights[hash]
		existingWeights.Add(existingWeights, weight)
	} else {
		o.Data[hash] = data
		o.Weights[hash] = weight
	}
	o.Total.Add(o.Total, weight)

}

/*
Scales all the existing weights on Odds.Weights by the given factor.
*/
func (o *Odds[D]) Scale(scaleFactor *big.Int) {
	for _, weight := range o.Weights {
		weight.Mul(weight, scaleFactor)
	}
	o.Total.Mul(o.Total, scaleFactor)
}

/*
Merges 2 odds of the same type together and produce a new Odds object that
represents the convolution of the two.
*/
func Combine[D Hashable[D]](
	o1 *Odds[D],
	o2 *Odds[D],
) *Odds[D] {
	newOdds := &Odds[D]{
		Data:    map[string]D{},
		Weights: map[string]*big.Int{},
		Total:   big.NewInt(0),
	}

	for h1, d1 := range o1.Data {
		w1 := o1.Weights[h1]

		for h2, d2 := range o2.Data {
			w2 := o2.Weights[h2]

			newData := d1.Combine(d2)
			newWeight := big.NewInt(0)
			newWeight.Mul(w1, w2)

			newOdds.Add(newData, newWeight)

		}
	}

	return newOdds
}

/////////////
// HELPERS //
/////////////

func (o *Odds[D]) String() string {

	s := "Total Weight: " + fmt.Sprint(o.Total) + "\n"

	hashes := []string{}
	longestHash := 0
	for _, data := range o.Data {
		dataHash := data.Hash()

		if len(dataHash) > longestHash {
			longestHash = len(dataHash)
		}

		hashes = append(hashes, dataHash)
	}

	sort.SliceStable(hashes, func(i, j int) bool {
		return hashes[i] < hashes[j]
	})

	for _, hash := range hashes {
		weight := o.Weights[hash]
		line := hash + (strings.Repeat(" ", longestHash-len(hash)))
		s += line + ": " + fmt.Sprint(weight) + "\n"
	}
	return s[:len(s)-1]
}

//////////////////
// CONSTRUCTORS //
//////////////////

/*
Create a new Odds object with arbitrary weight precision.
Requires a function that hashes the data types and a function which combines
two data types into a new data type.

Instantiate with "New[*{{DATA_TYPE}}]()"
*/
func New[D Hashable[D]]() *Odds[D] {
	return &Odds[D]{
		Data:    map[string]D{},
		Weights: map[string]*big.Int{},
		Total:   big.NewInt(0),
	}
}
