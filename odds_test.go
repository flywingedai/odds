package odds_test

import (
	"fmt"
	"math/big"
	"runtime"
	"testing"

	"github.com/flywingedai/odds"
)

func TestOptions(t *testing.T) {
	// x := 3
	// y := 4
	// z := 5

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

	// assert.Equal(t, 3, testOdds.HashFunction(&x))

	// assert.Equal(t, 5, *(testOdds.CopyFunction(&z)))

	// assert.Equal(t, 7, *(testOdds.CombineFunction(&x, &y)))

	// testOdds.CombineInPlaceFunction(&x, &z)
	// assert.Equal(t, 8, x)

	// entries := testOdds.ConvolveFunction(testOdds, testOdds.NewEntry(&y, big.NewInt(1)), testOdds.NewEntry(&z, big.NewInt(1)))
	// assert.Equal(t, 20, *entries[0].Data)

	// testOdds.ConvolveInPlaceFunction(testOdds, testOdds.NewEntry(&x, big.NewInt(1)), testOdds.NewEntry(&y, big.NewInt(1)))
	// assert.Equal(t, 32, x)

	// assert.Equal(t, "32", testOdds.DisplayFunction(&x))

	for i := 1; i <= 2_000; i++ {
		x := i
		testOdds.Add(&x, big.NewInt(int64(i)))
	}

	// testOdds.Extend_Parallel(func(e *odds.Entry[*int, int]) *int {
	// 	newValue := 0
	// 	for i := 0; i < 10_000; i++ {
	// 		newValue += *e.Data
	// 	}
	// 	return &newValue
	// }, 12)

	runtime.GOMAXPROCS(60)

	testOdds.ExtendOdds_Parallel(func(o *odds.Odds[*int, int], e *odds.Entry[*int, int]) *odds.Odds[*int, int] {
		newOdds := o.AsReference()

		for i := 1; i <= *e.Data+3; i++ {
			newData := i * *e.Data
			newOdds.Add(&newData, big.NewInt(1))
		}

		return newOdds
	}, 60, odds.Modify_Default)

	// .42 seconds
	// 64679 15380503786035895166827898308685180359996955436116372622880945780240311049137331822092315381576728395223808
	// 81817768248492198522608089418484814834832047021785344127742133953603490910800108573042030850503878217794670308392000000

	// .30 seconds
	// 64679 15380503786035895166827898308685180359996955436116372622880945780240311049137331822092315381576728395223808
	// 81817768248492198522608089418484814834832047021785344127742133953603490910800108573042030850503878217794670308392000000

	// testOdds.ExtendOdds(func(o *odds.Odds[*int, int], e *odds.Entry[*int, int]) *odds.Odds[*int, int] {
	// 	newOdds := o.AsReference()

	// 	for i := 1; i <= *e.Data+3; i++ {
	// 		newData := i * *e.Data
	// 		newOdds.Add(&newData, big.NewInt(1))
	// 	}

	// 	return newOdds
	// }, odds.Modify_Default)

	// 3.3

	fmt.Println(len(testOdds.Map), testOdds.Total)
	// for _, entry := range testOdds.EntriesByWeight() {
	// 	fmt.Println(entry)
	// }

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
