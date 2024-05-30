package odds

import "math/big"

/*
For every entry in "o", check if it satisties the "removalCondition" function.
If it does, that entry will be removed from "o"
*/
func (o *Odds[D, H]) RemoveCondition(removalCondition func(*Entry[D, H]) bool) *Odds[D, H] {
	for _, entry := range o.Entries() {
		if removalCondition(entry) {
			o.RemoveEntry(entry)
		}
	}
	return o
}

/*
For each condifionFunction passed in, create a new odds object of all the
entries that satisfy it. Earlier conditions have priority. The last returned
odds object is original odds with everything removed
*/
func (o *Odds[D, H]) SplitByConditions(
	conditionFunctions ...func(*Entry[D, H]) bool,
) []*Odds[D, H] {

	oddsArray := []*Odds[D, H]{}

	for i, conditionFunction := range conditionFunctions {
		oddsArray = append(oddsArray, NewOddsFromReference(o))
		for _, entry := range o.Entries() {
			if conditionFunction(entry) {
				oddsArray[i].AddEntry(entry)
				o.RemoveEntry(entry)
			}
		}
	}

	oddsArray = append(oddsArray, o)
	return oddsArray
}

/*
Returns the weight of entries in "o" which satisfy the given condition.
*/
func (o *Odds[D, H]) ConditionWeight(condition func(*Entry[D, H]) bool) *big.Int {
	count := big.NewInt(0)

	for _, entry := range o.Map {
		if condition(entry) {
			count.Add(count, entry.Weight)
		}
	}

	return count
}

// Returns true if all entries in the odds object satisfy the condition
func (o *Odds[D, H]) ConditonAllTrue(condition func(*Entry[D, H]) bool) bool {
	for _, entry := range o.Map {
		if !condition(entry) {
			return false
		}
	}
	return true
}

// Returns true if all entries in the odds object do not satisfy the condition
func (o *Odds[D, H]) ConditionAllFalse(condition func(*Entry[D, H]) bool) bool {
	for _, entry := range o.Map {
		if condition(entry) {
			return false
		}
	}
	return true
}
