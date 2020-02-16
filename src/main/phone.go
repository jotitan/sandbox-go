package main

import (
	"fmt"
	"math"
	"sort"
)

func main(){

	squares := findSquares()
	unions := make([]int,0)
	for i,s := range squares {
		if i == 0 {
			unions = findFactors(s)
		}else{
			unions = union(unions,findFactors(s))
		}
	}
	fmt.Println("UNION",unions)
}

var primesList = []int{2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,61,67,71,73,79,83,89,97}

func union(from, to []int)[]int{
	if len(from) == 0 || len(to) == 0 {
		return []int{}
	}
	sort.Ints(from)
	sort.Ints(to)
	posTo := 0
	posFrom := 0
	union := make([]int,0,len(from))
	for{
		if posTo >= len(to) || posFrom >= len(from){
			break
		}
		if from[posFrom] == to[posTo] {
			union = append(union,from[posFrom])
			posTo++
			posFrom++
		}else{
			if from[posFrom] > to[posTo]{
				posTo++
			}else{
				posFrom++
			}
		}
	}
	return union
}

func findSquares()[]int{
	data := [][]int{{7, 8, 9},{4,5,6},{1,2,3}}
	squares := make([]int,0)
	for i := 0 ; i < 2 ; i++ {
		for j := 0 ; j < 2 ; j++ {
			// Create list with square
			list := []int{data[i][j],data[i+1][j],data[i+1][j+1],data[i][j+1]}
			for combi := 0 ; combi < 4 ; combi++ {
				sumF := 0
				sumB := 0
				for r := 0 ; r < 4 ; r++{
					value := list[(r+combi)%4]
					sumF+=value*int(math.Pow(float64(10),float64(r)))
					sumB+=value*int(math.Pow(float64(10),float64(3-r)))
				}
				squares = append(squares,sumF,sumB)
			}
		}
	}
	return squares
}

func findFactors(val int)[]int{
	factors := make([]int,0,len(primesList))
	for _,v := range primesList {
		if val % v == 0 {
			factors = append(factors,v)
		}
	}
	return factors
}