package longhorn

import (
	"errors"
	"net/http"
)


// Manage workflow of game


type Event struct {
	CasePos int8
	// number of the player, 0 for the first
	Player int
	NextCasePos int
	Color int
	Info *InfoAction
}

// Behave of player
type InfoAction struct {
	TypeAction int
	Color int
	CasePos int
	// for each color, position of the case
	Cases map[string]int
}

func checkEvent(b Board,e Event)error {
	if b.currentPlayer.id != e.Player {
		return errors.New("Problem with player")
	}
	if b.currentCase.position != e.CasePos {
		return errors.New("Probleme with event")
	}
	if (e.Info != nil && !b.currentCase.PlayAction()) ||
			(e.Info == nil && b.currentCase.PlayAction()) {
		return errors.New("An action must be played")
	}
	if e.Info != nil && e.Info.TypeAction != b.currentCase.action.actionType {
		return errors.New("Action type must be same")
	}
	return nil
}

func workflow(b Board,e Event)error{
	if err := checkEvent(b,e) ; err != nil {
		return err
	}
	// execute standard action
	nbCow := b.currentCase.TakeColor(e.Color)
	b.currentPlayer.GetCows(map[int]int8{e.Color:nbCow})

	// execute specific action if necessary
	switchPlayer := true
	if e.Info != nil {
		switchPlayer = b.currentCase.ExecuteSpecificAction(b,*e.Info)
	}
	// Move next player
	b.MoveCase(e.NextCasePos)

	// Switch player or not
	if switchPlayer {
		 b.SwitchPlayer()
	}

	// Send data to user (player info and board data)
//	m := NewServerMessage(b)


	// Check if a winner
	return nil
}

// PlayerDialog manage communication from server to player (SSE)
type PlayerDialog struct{
	response http.ResponseWriter
}

func (pd * PlayerDialog)createSSEHeader(){
	pd.response.Header().Set("Content-Type","text/event-stream")
	pd.response.Header().Set("Cache-Control","no-cache")
	pd.response.Header().Set("Connection","keep-alive")
	pd.response.Header().Set("Access-Control-Allow-Origin","*")
}

func (pd PlayerDialog)sendMessage(data []byte){
	pd.response.Write([]byte("data: " + string(data) + "\n\n"))
	pd.response.(http.Flusher).Flush()
}

type Game struct{
	Board Board

}


type GameManager struct{
	games []*Game
	gamesById map[int]*Game
	counterIdGame int
}

func NewGameManager()GameManager{
	return GameManager{make([]*Game,0),make(map[int]*Game),0}
}

func (gm GameManager)GetGame(idGame int)(*Game,error){
	if game,exist:= gm.gamesById[idGame] ; exist{
		return game,nil
	}
	return nil,errors.New("Game doesn't exist")
}

func (gm * GameManager)CreateGame()*Game{
	id := gm.counterIdGame
	gm.counterIdGame++
	g := Game{NewBoard(-1,id)}
	gm.games = append(gm.games,&g)
	gm.gamesById[id] = &g
	return &g
}
