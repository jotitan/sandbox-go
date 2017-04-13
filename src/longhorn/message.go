package longhorn

import (
	"fmt"
	"encoding/json"
)


// manage message between server and client

type CaseMessage struct{
	Name string
	Action string
	// money case, value of money
	ActionInfo string
	Nb int8
	Position int8
	Cows []int
}

func newCaseMessage(c Case)CaseMessage {
	cm := CaseMessage{Name:c.name,Nb:c.cowNumber,Position:c.position,Cows:c.cows}
	if c.action!=nil {
		cm.Action = c.action.Name()
		if cm.Action == "money"{
			cm.ActionInfo = fmt.Sprintf("$ %d",c.action.action.(MoneyAction).money)
		}
	}
	return cm
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
	// if an action must be played
	Action string
	// Option to action
	Neighbors []int
	// colors when snake
	Colors []int
	// Save previous winner
	Winner int
	Info string
}

func (sm ServerMessage)ToJSON()[]byte{
	str,_ := json.Marshal(sm)
	return str
}

func NewBeginServerMessage(b Board)ServerMessage{
	sm := NewServerMessage(b)
	// Set neighbors with case with 4 elements on it
	possiblesCases := make([]int,0,5)
	for _,c := range b.cases {
		if c.cowNumber == 4 {
			possiblesCases = append(possiblesCases,int(c.position))
		}
	}
	sm.Neighbors = possiblesCases
	sm.Action = "begin"
	return sm
}

func NewWinServerMessage(b Board,winner Player, cause string)ServerMessage{
	sm := NewBeginServerMessage(b)
	sm.Winner= winner.Id
	sm.Info = cause
	sm.Action = "win"
	return sm
}

func NewMoveServerMessage(b Board, moves []int)ServerMessage{
	sm := NewServerMessage(b)
	sm.Action = "move"
	// Move possible on position 0
	sm.Moves["0"] = moves
	return sm
}

func NewServerMessage(b Board)ServerMessage{
	// cases reachable for each color (based on number of cow)
	sm := ServerMessage{}
	sm.NextPlayer = b.currentPlayer.Id
	sm.P1 = *b.p1
	sm.P2 = *b.p2
	sm.Id = b.idBoard

	sm.Moves = make(map[string][]int)
	if b.currentCase != nil {
		for color, nb := range b.currentCase.cows {
			if nb > 0 {
				sm.Moves[fmt.Sprintf("%d", color)] = b.FindPlayableCases(int(b.currentCase.position), nb)
			}
		}
	}
	sm.CurrentCase = int8(-1)
	if b.currentCase!= nil {
		sm.CurrentCase = b.currentCase.position
	}
	sm.Cases = make([]CaseMessage,len(b.cases))
	for i,c := range b.cases {
		sm.Cases[i] = newCaseMessage(*c)
	}

	// if only one color stayed, set action
	if b.currentCase!=nil {
		nbColor := 0
		for _, nb := range b.currentCase.cows {
			if nb > 0 {
				nbColor++
			}
		}
		if nbColor == 1 && b.currentCase.action!=nil{
			sm.Action = b.currentCase.action.Name()
			switch sm.Action {
				case "swallow" :
				nbs := b.GetNeighbors(int(b.currentCase.position))
				sm.Neighbors = make([]int,len(nbs))
				for i,c := range nbs {
					sm.Neighbors[i] = int(c.position)
				}
				case "snake" :
				// Define color that have to be dispatch (color on case and color of user)
				sm.Colors = make([]int,4)
				for color,nb := range b.currentPlayer.Cows {
					if nb > 0 {
						sm.Colors[color] = 1
					}
				}
				for color,nb := range b.currentCase.cows {
					if nb > 0 {
						sm.Colors[color] = 1
					}
				}
				sm.Neighbors = make([]int,0)
				for _,c := range b.GetNeighbors(int(b.currentCase.position)){
					sm.Neighbors = append(sm.Neighbors,int(c.position))
				}
			}
		}
	}
	return sm
}
