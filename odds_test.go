package odds_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/flywingedai/odds"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	x := 3
	y := 4
	z := 5

	testOdds := odds.NewOptions(test_HashFunction1).Odds()
	// assert.Equal(t, 6, testOdds.HashFunction(&x))

	testOdds = odds.NewOptions(test_HashFunction1).
		WithHash(test_HashFunction2).
		WithCopy(test_CopyFunction).
		WithAdd(test_CombineFunction).
		WithAddInPlace(test_CombineInPlaceFunction).
		WithConvolve(test_ConvolveFunction).
		WithConvolveInPlace(test_ConvolveInPlaceFunction).
		WithDisplay(test_DisplayFunction).Odds()

	smallTests := testOdds.Copy()
	assert.Equal(t, 3, smallTests.HashFunction(&x))

	assert.Equal(t, 5, *(smallTests.CopyFunction(&z)))

	assert.Equal(t, 7, *(smallTests.CombineFunction(&x, &y)))

	smallTests.CombineInPlaceFunction(&x, &z)
	assert.Equal(t, 8, x)

	entries := smallTests.ConvolveFunction(smallTests, smallTests.NewEntry(&y, big.NewInt(1)), smallTests.NewEntry(&z, big.NewInt(1)))
	assert.Equal(t, 20, *entries[0].Data)

	smallTests.ConvolveInPlaceFunction(smallTests, smallTests.NewEntry(&x, big.NewInt(1)), smallTests.NewEntry(&y, big.NewInt(1)))
	assert.Equal(t, 32, x)

	assert.Equal(t, "32", smallTests.DisplayFunction(&x))

	for i := 1; i <= 5; i++ {
		x := i
		testOdds.Add(&x, big.NewInt(int64(i)))
	}

	extend := testOdds.Copy()
	extend.Extend(func(e *odds.Entry[*int, int]) *int {
		newValue := 0
		for i := 0; i < 10_000; i++ {
			newValue += *e.Data
		}
		return &newValue
	})

	extendParallel := testOdds.Copy()
	extendParallel.Extend_Parallel(func(e *odds.Entry[*int, int]) *int {
		newValue := 0
		for i := 0; i < 10_000; i++ {
			newValue += *e.Data
		}
		return &newValue
	}, 4)

	extendOdds := testOdds.Copy()
	extendOdds.ExtendOdds(func(o *odds.Odds[*int, int], e *odds.Entry[*int, int]) *odds.Odds[*int, int] {
		newOdds := o.AsReference()

		for i := 1; i <= *e.Data+3; i++ {
			newData := i * *e.Data
			newOdds.Add(&newData, big.NewInt(1))
		}

		return newOdds
	}, odds.Add_Combine)

	extendOdds = testOdds.Copy()
	extendOdds.ExtendOdds(func(o *odds.Odds[*int, int], e *odds.Entry[*int, int]) *odds.Odds[*int, int] {
		newOdds := o.AsReference()

		for i := 1; i <= *e.Data+3; i++ {
			newData := i * *e.Data
			newOdds.Add(&newData, big.NewInt(1))
		}

		return newOdds
	}, odds.Add_Default)

	extendOddsParallel := testOdds.Copy()
	extendOddsParallel.ExtendOdds_Parallel(func(o *odds.Odds[*int, int], e *odds.Entry[*int, int]) *odds.Odds[*int, int] {
		newOdds := o.AsReference()

		for i := 1; i <= *e.Data+3; i++ {
			newData := i * *e.Data
			newOdds.Add(&newData, big.NewInt(1))
		}

		return newOdds
	}, 4, odds.Add_CombineInPlace)

}

// Test Functions //

func test_HashFunction1(i *int) int { return 2 * (*i) }
func test_HashFunction2(i *int) int { return *i }
func test_CopyFunction(i *int) *int {
	x := *i
	return &x
}
func test_CombineFunction(i1, i2 *int) *int {
	x := *i1 + *i2
	return &x
}
func test_CombineInPlaceFunction(i1, i2 *int) {
	*i1 = *i1 + *i2
}
func test_ConvolveFunction(o *odds.Odds[*int, int], i1, i2 *odds.Entry[*int, int]) []*odds.Entry[*int, int] {
	x := *i1.Data * *i2.Data
	return []*odds.Entry[*int, int]{o.NewEntry(&x, big.NewInt(1))}
}
func test_ConvolveInPlaceFunction(_ *odds.Odds[*int, int], i1, i2 *odds.Entry[*int, int]) {
	*i1.Data = (*i1.Data) * (*i2.Data)
}
func test_DisplayFunction(i *int) string { return fmt.Sprint(*i) }
