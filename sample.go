package odds

import (
	"crypto/rand"
	"math/big"
)

/*
Get a single random sample from the odds object. Returns the full entry.
*/
func (o *Odds[D, H]) Sample() *Entry[D, H] {
	total := big.NewInt(0)
	randPoint, _ := rand.Int(rand.Reader, o.Total)
	for _, entry := range o.Map {
		total.Add(total, entry.Weight)
		if total.Cmp(randPoint) >= 0 {
			return entry
		}
	}
	return nil
}
