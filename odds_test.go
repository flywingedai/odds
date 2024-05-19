package odds

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData int

func (d *testData) Hash() string {
	return fmt.Sprint(*d)
}

func (d *testData) Combine(other *testData) *testData {
	result := *d + *other
	return &result
}

func Test(t *testing.T) {
	m := New[*testData]()
	for i := 1; i <= 10; i++ {
		d := testData(i)
		m.Add(&d, big.NewInt(int64(10*i)))
	}
	m.Scale(big.NewInt(2))
	_ = fmt.Sprint(m)
}

func TestDuplicateAdd(t *testing.T) {
	m := New[*testData]()
	for i := 1; i <= 10; i++ {
		d := testData(5)
		m.Add(&d, big.NewInt(int64(10*i)))
	}
}

func TestCombine(t *testing.T) {
	m1 := New[*testData]()
	m2 := New[*testData]()
	for i := 1; i <= 10; i++ {
		d1 := testData(i)
		m1.Add(&d1, big.NewInt(int64(1)))

		d2 := testData(i)
		m2.Add(&d2, big.NewInt(int64(1)))
	}

	assert.Equal(t, "10", m1.Total.String())
	assert.Equal(t, "10", m2.Total.String())

	m3 := Combine(m1, m2)
	assert.Equal(t, "100", m3.Total.String())
}
