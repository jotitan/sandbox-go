package longhorn

import (
	"math/rand"
	"sort"
)


// Used to sort randomly list of element

type randomCouples struct{
	// list to sort
	values []int
	// random list to sort. Same number of element than values
	randomValues []int
}

func(r randomCouples)Len() int{return len(r.values)}
func(r randomCouples)Less(i, j int) bool {
	return r.randomValues[i] < r.randomValues[j]
}
func(r randomCouples)Swap(i, j int){r.values[i],r.values[j] = r.values[j],r.values[i]}

func sortRandomly(values []int){
	randomValues := make([]int,len(values))
	for i := range values {
		randomValues[i] = rand.Int()
	}
	sort.Sort(randomCouples{values,randomValues})
}
