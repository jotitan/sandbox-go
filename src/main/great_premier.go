package main

import (
	"fmt"
	"math"
	"time"
)

func main(){
	results := int64(0)
	begin := time.Now()
	searchGreat(1,0,&results)

	fmt.Println("GOT ",results,"in",time.Now().Sub(begin))
}

func searchGreat(pow10, current int64, results *int64)bool{
	findOne := false
	if pow10 <= 10000000000 {
		for i := int64(0) ; i < 10 ; i++ {
			value := current + i * pow10
			if isPremier(value){
				if !searchGreat(pow10*10,value,results){
					if value > *results {
						*results = value
					}
				}
				findOne = true
			}
		}
	}
	return findOne
}

func isPremier(value int64)bool{
	if value%2 == 0 {
		return false
	}
	for i := int64(3) ; i < int64(math.Sqrt(float64(value)))+2 ; i+=2 {
		if value % i == 0 {
			return false
		}
	}
	return true
}
