package main

import (
	"longhorn"
	"fmt"
	"encoding/json"
)


func main(){
	b := longhorn.NewBoard(-1)
	b.GetNeighbors(2)
	b.GetNeighbors(4)
	b.GetNeighbors(5)
	b.GetNeighbors(6)
	b.GetNeighbors(7)
	b.GetNeighbors(8)

	//fmt.Println(b.Find(2,1,-1))
	fmt.Println(b.FindCases(1,4,-1))


	e := longhorn.Event{Info:&longhorn.InfoAction{}}
	d,err := json.Marshal(e)
	fmt.Println(err,string(d))

	sm := longhorn.NewServerMessage(b)
	fmt.Println(sm)

}
