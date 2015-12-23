package longhorn

import "fmt"


// manage message between server and client

type CaseMessage struct{
	Name string
	Nb int8
	Position int8
	Cows []int
}

func newCaseMessage(c Case)CaseMessage {
	return CaseMessage{c.name,c.cowNumber,c.position,c.cows}
}

type ServerMessage struct{
	// info about player status, game status
	NextPlayer int
	// case where player are
	CurrentCase int8
	// action to play (if only one color stayed on case)
	Moves map[string][]int
	// info about player
	P1 Player
	P2 Player
	Cases []CaseMessage
	Id int
}

func NewServerMessage(b Board)ServerMessage{
	// cases reachable for each color (based on number of cow)
	moves := make(map[string][]int)
	if b.currentCase != nil {
		for color, nb := range b.currentCase.GetCowsByColor() {
			moves[fmt.Sprintf("%d",color)] = b.FindPlayableCases(int(b.currentCase.position), nb)
		}
	}
	casePosition := int8(-1)
	if b.currentCase!= nil {
		casePosition = b.currentCase.position
	}
	cases := make([]CaseMessage,len(b.cases))
	for i,c := range b.cases {
		cases[i] = newCaseMessage(*c)
	}
	// Add all case info or only delta ?
	return ServerMessage{b.currentPlayer.id,casePosition,moves,b.p1,b.p2,cases,b.idBoard}
}
