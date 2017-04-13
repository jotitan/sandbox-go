package longhorn

import (
	"errors"
	"net/http"
	"fmt"
	"encoding/json"
	"logger"
)


// Manage workflow of game


type Event struct {
	CasePos int8
	// number of the player, 0 for the first
	Player int
	NextCasePos int
	Color int
	GameId int
	Info *InfoAction
}

// Behave of player
type InfoAction struct {
	TypeAction int
	Color int
	CasePos int
	// for each color, position of the case (for snake)
	Cases map[string]int
}

func(ia *InfoAction)String()string{
	return fmt.Sprintf("Color %d, case %d",ia.Color,ia.CasePos)
}

func EventFromJSON(data []byte)(Event,error){
	e := Event{}
	if err := json.Unmarshal(data,&e) ; err != nil {
		return Event{},err
	}
	return e,nil
}

func checkEvent(b Board,e Event)error {
	if b.currentPlayer.Id != e.Player {
		return errors.New("Problem with player")
	}
	if b.currentCase.position != e.CasePos {
		return errors.New("Probleme with event")
	}
	if (e.Info != nil && !b.currentCase.PlayAction()) ||
			(e.Info == nil && b.currentCase.PlayAction()) {
		return errors.New("An action must be played")
	}

	/*if e.Info != nil && e.Info.TypeAction != b.currentCase.action.actionType {
		return errors.New("Action type must be same")
	}*/
	return nil
}

func (g * Game)begin()error{
	// Define first player
	g.Board.currentPlayer = g.Board.p1
	beginMessage := NewBeginServerMessage(g.Board)
	beginMoves := make([]int,0,5)
	for _,c := range g.Board.cases {
		if c.cowNumber == 4 {
			beginMoves = append(beginMoves,int(c.position))
		}
	}
	beginMessage.Moves["0"] = beginMoves
	g.SendToAll("",beginMessage.ToJSON())
	return nil
}

func (g * Game)SendInfo()error{
	if g.Board.currentCase == nil{
		return g.begin()
	}
	var m []byte
	if g.Board.currentCase.action == nil && !g.Board.currentCase.HasCows(){
		// move case
		// Use previous cowNumber get
		moves := g.Board.FindPlayableCases(int(g.Board.currentCase.position),g.Board.previousCowNumber)
		m = NewMoveServerMessage(g.Board,moves).ToJSON()
	}else {
		m = NewServerMessage(g.Board).ToJSON()
	}
	g.SendToAll("",m)
	return nil
}

// When a player win the game. Send message, reinit the board
func (g *Game)SendWinner(winner * Player,cause string)error{
	if winner == nil {
		// no winner
	}
	winner.Point++

	g.Board.ResetBoard(g.previousFirstPlayer)
	g.previousFirstPlayer = g.Board.currentPlayer.Id

	m := NewWinServerMessage(g.Board,*winner,cause).ToJSON()
	g.SendToAll("",m)

	return nil
}

func (g * Game)Workflow(e Event)error{
	logger.GetLogger().Info("Receive event",e)
	if e.GameId == 0 && g.Board.currentCase == nil{
		return g.begin()
	}

	// execute standard action. Not executed when begin case
	nbCow := int8(0)
	if e.Color!=-1 {
		if err := checkEvent(g.Board,e) ; err != nil {
			logger.GetLogger().Error(err)
			return err
		}
		nbCow = g.Board.currentCase.TakeColor(e.Color)
		g.Board.currentPlayer.GetCows(map[int]int8{e.Color:nbCow})
		g.Board.previousCowNumber = int(nbCow)
	}

	// execute specific action if necessary
	switchPlayer := true
	moveCase := true
	if e.Info != nil {
		switchPlayer,moveCase = g.Board.currentCase.ExecuteSpecificAction(g.Board,*e.Info)
	}
	var m []byte
	// check win : no move available, 9 of one color, sheriff
	if hasWinner,winner := g.HasWinner() ; hasWinner {
		return g.SendWinner(winner, "get 9 of one color")
	}else {
		// Some action avoid moving case cause board is changing (swallow, snake, killcolor)
		// Move next player

		// If no move available, win. Warning when move case (when nbCow = 0)
		if g.Board.currentCase!=nil && nbCow > 0{
			if moves := g.Board.FindPlayableCases(int(g.Board.currentCase.position), int(nbCow)) ; len(moves) == 0 {
				winner, score1,score2 := g.Board.FindWinner()
				return g.SendWinner(winner,fmt.Sprintf("No move, win by score : %d / %d",score1,score2))
			}
		}
		if moveCase {
			// If no move, end, compute results
			g.Board.MoveCase(e.NextCasePos)
			// Switch player or not
			if switchPlayer {
				g.Board.SwitchPlayer()
			}
			// Test sheriff case
			if g.Board.currentCase.action!=nil && g.Board.currentCase.action.IsSheriff() && g.Board.currentCase.PlayAction() {
				// Player current lose
				return g.SendWinner(g.Board.otherPlayer, "looser got a sheriff")
			}
			m = NewServerMessage(g.Board).ToJSON()
		}else {
			m = NewMoveServerMessage(g.Board, g.Board.FindPlayableCases(int(g.Board.currentCase.position), int(nbCow))).ToJSON()
		}

	}
	// Send data to user (player info and board data)

	g.SendToAll("",m)
	return nil
}

// PlayerDialog manage communication from server to player (SSE)
type PlayerDialog struct{
	response http.ResponseWriter
	sessionId string
}

func (pd * PlayerDialog)setResponse(response http.ResponseWriter){
	pd.response = response
	pd.createSSEHeader()
}

func (pd * PlayerDialog)createSSEHeader(){
	pd.response.Header().Set("Content-Type","text/event-stream")
	pd.response.Header().Set("Cache-Control","no-cache")
	pd.response.Header().Set("Connection","keep-alive")
	pd.response.Header().Set("Access-Control-Allow-Origin","*")
}

func (pd PlayerDialog)sendMessage(event string,data []byte){
	if event != "" {
		pd.response.Write([]byte(fmt.Sprintf("event: %s\n",event)))
	}
	pd.response.Write([]byte("data: " + string(data) + "\n\n"))
	pd.response.(http.Flusher).Flush()
}

func (pd PlayerDialog)IsConnected()bool{
	return true
}

type Game struct{
	Board Board
	pd1 *PlayerDialog
	pd2 *PlayerDialog
	previousFirstPlayer int
}

func (g Game)HasWinner()(bool,*Player){
	if g.Board.p1.IsWinner() {
		return true,g.Board.p1
	}
	if g.Board.p2.IsWinner() {
		return true,g.Board.p2
	}
	return false,nil
}

func (g * Game)ConnectPlayer(r http.ResponseWriter,sessionId string)(*PlayerDialog,error){
	if g.pd1 == nil && g.pd2 == nil {
		return nil,errors.New("No player recorded on game")
	}
	var p * PlayerDialog
	id := 1
	if g.pd1 != nil && g.pd1.sessionId == sessionId {
		p = g.pd1
	}
	if g.pd2 != nil && g.pd2.sessionId == sessionId {
		p = g.pd2
		id = 2
	}
	if p != nil {
		p.setResponse(r)
		p.sendMessage("info",[]byte("You are connected"))
		p.sendMessage("userid",[]byte(fmt.Sprintf("%d",id)))
		if g.pd1 != nil && g.pd2 != nil {
			// Run game
			g.SendInfo()
		}
		return p,nil
	}
	return nil,errors.New("Impossible to connect")
}

func (g Game)SendToAll(event string,data []byte)error{
	logger.GetLogger().Info("TOALL",string(data))
	g.pd1.sendMessage(event,data)
	g.pd2.sendMessage(event,data)
	return nil
}

type GameManager struct{
	games []*Game
	gamesById map[int]*Game
	counterIdGame int
}

func NewGameManager()GameManager{
	return GameManager{make([]*Game,0),make(map[int]*Game),1}
}

// GetGame return game and record player at empty place. If no place, check if sessionIdis the same
func (gm GameManager)GetGame(idGame int,sessionId string, name string)(*Game,error){
	if game,exist:= gm.gamesById[idGame] ; exist{
		if (game.pd1!= nil && game.pd1.sessionId == sessionId) ||
		(game.pd2!=nil && game.pd2.sessionId == sessionId) {
			return game, nil
		}
		if game.pd1 == nil {
			game.pd1 = &PlayerDialog{sessionId:sessionId}
			if name != "" {
				game.Board.p1.Name = name
			}
			return game, nil
		}
		if game.pd2 == nil {
			game.pd2 = &PlayerDialog{sessionId:sessionId}
			if name != "" {
				game.Board.p2.Name = name
			}
			return game, nil
		}
		return nil,errors.New("No place available")
	}
	return nil,errors.New("Game doesn't exist")
}

func (gm * GameManager)CreateGame(sessionId string, name string)*Game{
	id := gm.counterIdGame
	gm.counterIdGame++
	g := Game{Board:NewBoard(id)}
	g.previousFirstPlayer = g.Board.currentPlayer.Id
	g.pd1 = &PlayerDialog{sessionId:sessionId}
	g.Board.p1.Name = name
	gm.games = append(gm.games,&g)
	gm.gamesById[id] = &g
	return &g
}
