package longhorn

import (
	"fmt"
	"strconv"
	"math/rand"
)

type Color int

const (
	blackColor = 0
	orangeColor = 1
	greenColor = 2
	whiteColor = 3

	typeSnakeAction = 0
	typeSheriffAction = 1
	typeMoney200Action = 2
	typeMoney300Action = 3
	typeMoney400Action = 4
	typeMoney500Action = 5
	typeReplayAction = 6
	typeKillColorAction = 7
	typeStoleAction = 8
	typeSwallowAction = 9
)

var caseNames = []string{"Nugget Hill","Tulsa","Salomon Tree","Kid Copper's Ranch","Red River Valley","Maverick Junction","Cherokee Spring","Dagger Flat","Blackstone Corral"}
var caseNumbers = []int8{6,2,3,4,4,4,4,4,5}
var actions = []int{
	typeSnakeAction,typeSnakeAction,typeSheriffAction,typeKillColorAction,
	typeMoney200Action,typeMoney300Action,typeStoleAction,typeMoney400Action,typeMoney400Action,typeMoney500Action,
	typeReplayAction,typeReplayAction,typeReplayAction,
	typeSwallowAction,typeSwallowAction,typeSwallowAction,
	typeStoleAction,typeMoney300Action,typeStoleAction,
}

type Board struct{
	cases [9]*Case
	casesByPosition [9]*Case
	p1 * Player
	p2 * Player
	currentPlayer * Player
	otherPlayer * Player
	currentCase * Case
	idBoard int
	previousCowNumber int
}

func (b Board)GetId()int{
	return b.idBoard
}

func (b Board)GetOtherPlayer()*Player{
	return b.otherPlayer
}

func (b * Board)KillColor(color int){

	// kill color for every player
	b.p1.LoseCows(map[int]int8{color:9})
	b.p2.LoseCows(map[int]int8{color:9})

	for _,c := range b.cases {
		if c.TakeColor(color) > 0 && !c.HasCows(){
			// remove action if case empty
			c.action = nil
		}
	}
}

func (b * Board)MoveCase(nextCase int){
	b.currentCase = b.casesByPosition[nextCase]
}

func (b * Board)SwitchPlayer(){
	if b.currentPlayer == nil {
		// Select first
		b.currentPlayer = b.p1
		b.otherPlayer = b.p2
		return
	}
	if b.currentPlayer.Id == b.p1.Id {
		b.currentPlayer = b.p2
		b.otherPlayer = b.p1
	}else{
		b.currentPlayer = b.p1
		b.otherPlayer = b.p2
	}
}

func (b Board)GetNeighbors(position int)[]*Case {
	casesId := make([]int,0)
	if position>2 {
		casesId = append(casesId,position-3)
	}
	if position%3 != 0 {
		casesId = append(casesId,position-1)
	}
	if position%3 != 2 {
		casesId = append(casesId,position+1)
	}
	if  position<6 {
		casesId = append(casesId,position+3)
	}
	cases := make([]*Case,len(casesId))
	for i,c := range casesId {
		cases[i] = b.casesByPosition[c]
	}
	return cases
}

var graph = map[int][]int{
	0:[]int{1,3},
	1:[]int{0,2,4},
	2:[]int{1,5},
	3:[]int{0,4,6},
	4:[]int{1,3,5,7},
	5:[]int{2,4,8},
	6:[]int{3,7},
	7:[]int{4,6,8},
	8:[]int{5,7},
}

// cases must contains cows
func (b Board)FindPlayableCases(root,deep int)[]int{
	if deep <=0 {
		return []int{}
	}
	positions := b.FindCases(root,deep,-1)
	filteredPositions := make([]int,0,len(positions))
	for _,pos := range positions {
		if b.casesByPosition[pos].HasCows() {
			filteredPositions = append(filteredPositions,pos)
		}
	}
	return filteredPositions
}

func (b Board)FindCases(root,deep int,previous int)[]int{
	if deep == 1 {
		results := make([]int,0,len(graph[root])-1)
		for _,pos := range graph[root] {
			if pos!=previous{
				results = append(results,pos)
			}
		}
		return results
	}
	mapResults := make(map[int]struct{})
	for _,e := range graph[root]{
		if e != previous {
			for _,pos := range b.FindCases(e,deep-1, root) {
				mapResults[pos] = struct{}{}
			}
		}
	}
	results := make([]int,0,len(mapResults))
    for pos := range mapResults {
 		results = append(results,pos)
	}
	return results
}

func (b Board)String()string{
	return fmt.Sprintf("%v",b.cases)
}

func getCows()[]int{
	cows := make([]int,36)
	for i := 0 ; i < len(cows) ; i++ {
		cows[i] = i%4
	}
	sortRandomly(cows)
	return cows
}

func (b Board)getRestCows()[]int{
	cows := make([]int,4)
	for _,c := range b.cases {
		for color,nb := range c.cows {
			cows[color]+=nb
		}
	}
	return cows
}

func (b Board)FindWinner()(*Player,int,int){
	restCows := b.getRestCows()
	score1,score2 := b.p1.CountPoint(restCows),b.p2.CountPoint(restCows)
	if score1 == score2{
		// No winner
		return nil ,0,0
	}
	if score1 > score2 {
		return b.p1,score1,score2
	}
	return b.p2,score2,score1
}

func (b * Board)selectPlayer(){
	b.p1 = &Player{Id:1}
	b.p2 = &Player{Id:2}

	if rand.Int()%2 == 0{
		b.currentPlayer = b.p1
		b.otherPlayer = b.p2
	}else{
		b.currentPlayer = b.p2
		b.otherPlayer= b.p1
	}
}

func (b * Board)initBoard(){
	positions := []int{0,1,2,3,4,5,6,7,8}
	cows := getCows()
	sortRandomly(positions)
	sortRandomly(actions)
	// If sheriff, must be on case 6. Do after loop and invert 2 cases
	posSheriff := -1
	posSixCases := -1
	for i,pos := range positions {
		nbCows := caseNumbers[i]
		if nbCows == 6{
			posSixCases = i
		}
		if actions[i] == typeSheriffAction {
			posSheriff = i
		}
		c := NewCase(caseNames[i],int8(nbCows),int8(pos),cows[:nbCows],actions[i])
		b.cases[i] = c
		b.casesByPosition[c.position] = c
		cows = cows[nbCows:]
	}
	// Sheriff case :
	if posSheriff != -1 && posSheriff != posSixCases {
		b.cases[posSixCases].action,b.cases[posSheriff].action =b.cases[posSheriff].action,b.cases[posSixCases].action
	}
	b.currentCase = nil
	// Rest player
	b.p1.Reset()
	b.p2.Reset()
}

func (b * Board)ResetBoard(firstPlayer int){
	b.initBoard()
	if (firstPlayer+1)%2 == 0 {
		b.currentPlayer,b.otherPlayer = b.p1,b.p2
	}else{
		b.currentPlayer,b.otherPlayer = b.p2,b.p1
	}
}

func NewBoard(idBoard int)Board{
	board := Board{idBoard:idBoard}
	board.selectPlayer()
	board.initBoard()
	return board
}

type Case struct {
	name string
	// number of cow
	cowNumber int8
	// Position on the board
	position int8
	// repartition cow
	cows []int
	// action when last cow is keep
	action * Action
}

func (c * Case)ExecuteSpecificAction(b Board,info InfoAction)(bool,bool){
	switchPlayer := true
	canMove := true
	if c.action != nil {
		switchPlayer = c.action.Do(b,info)
		canMove = c.action.CanMove()
		// action can only be executed one time
		c.action = nil
	}
	return switchPlayer,canMove
}

// PlayAction return true if action have to be played (only one color)
func (c Case)PlayAction()bool{
	nbColors := 0
	for _,nb := range c.cows {
		if nb > 0 {
			nbColors++
		}
	}
	return nbColors == 1
}

func (c Case)HasCows()bool{
	for _,nb := range c.cows {
		if nb > 0 {
			return true
		}
	}
	return false
}

// return nb of taken cows
func (c * Case)TakeColor(color int)int8{
	if color > len(c.cows){
		return 0
	}
	nb := c.cows[color]
	c.cows[color] = 0
	return int8(nb)
}

func (c * Case)AddColor(color,nb int){
	if color > len(c.cows){
		return
	}
	c.cows[color]+= nb
}

func (c Case)String()string{
	return fmt.Sprintf("%s (%d) at pos %d.\n",c.name,c.cowNumber,c.position)
}

func NewCase(name string,nb,pos int8,cows []int,typeAction int)*Case{
	// Create cows by color
	cowsByColor := make([]int,4)
	for _,c := range cows {
		cowsByColor[c]++
	}
	return &Case{name,nb,pos,cowsByColor,NewAction(typeAction)}
}

func NewAction(actionType int)*Action{
	var a CaseAction
	switch actionType {
		case typeMoney200Action : a = MoneyAction{200}
		case typeMoney300Action : a = MoneyAction{300}
		case typeMoney400Action : a = MoneyAction{400}
		case typeMoney500Action : a = MoneyAction{500}
		case typeKillColorAction : a = KillColorAction{}
		case typeReplayAction : a = ReplayAction{}
		case typeStoleAction : a = StoleAction{}
		case typeSwallowAction : a = SwallowAction{}
		case typeSheriffAction : a = SheriffAction{}
		case typeSnakeAction : a = SnakeAction{}
	}
	return &Action{actionType,a}
}

type Action struct{
	actionType int
	action CaseAction
}

func (a Action)Name()string{
	return a.action.Name()
}

func (a Action)IsSheriff()bool{
	return a.actionType == typeSheriffAction
}

// return true if player must change
func (a * Action)Do(b Board,info InfoAction)bool {
	return a.action.Do(b,info)
}

func (a * Action)CanMove()bool {
	return a.action.CanMove()
}

type CaseAction interface {
	Do(b Board,info InfoAction)bool
	CanMove()bool
	Name()string
}

type SnakeAction struct{}
func (sa SnakeAction)Name()string {
	return "snake"
}
func (sa SnakeAction)CanMove()bool{return false}
func (sa SnakeAction)Do(b Board,info InfoAction)bool {
	b.currentPlayer.LoseCows(map[int]int8{0:1,1:1,2:1,3:1})
	// dispatch cows on others cases

	for color,casePos := range info.Cases {
		if idColor,err := strconv.ParseInt(color,10,32) ; err == nil {
			b.casesByPosition[casePos].AddColor(int(idColor),1)
		}
	}

	return true
}

type SheriffAction struct{}
func (sa SheriffAction)Name()string{
	return "sheriff"
}
func (sa SheriffAction)CanMove()bool{return false}
func (sa SheriffAction)Do(b Board,info InfoAction)bool {
	return true
}

type MoneyAction struct{
	money int
}
func (sa MoneyAction)Name()string{
	return "money"
}
func (sa MoneyAction)CanMove()bool{return true}
func (sa MoneyAction)Do(b Board,info InfoAction)bool {
	b.currentPlayer.WinMoney(sa.money)
	return true
}

type ReplayAction struct{}
func (sa ReplayAction)Name()string{
	return "replay"
}
func (sa ReplayAction)CanMove()bool{return true}
func (sa ReplayAction)Do(b Board,info InfoAction)bool {
	return false
}

type KillColorAction struct{}
func (sa KillColorAction)Name()string{
	return "killcolor"
}
func (sa KillColorAction)CanMove()bool{return false}
func (sa KillColorAction)Do(b Board,info InfoAction)bool {
	b.KillColor(info.Color)
	return true
}

type StoleAction struct{}
func (sa StoleAction)Name()string{
	return "stole"
}
func (sa StoleAction)CanMove()bool{return true}
func (sa StoleAction)Do(b Board,info InfoAction)bool {
	if info.Color!= -1 {
		stolenCows := b.GetOtherPlayer().LoseCows(map[int]int8{info.Color:2})
		b.currentPlayer.GetCows(stolenCows)
	}else{
		// Lose higher money
		if money := b.GetOtherPlayer().LoseMoney() ; money > 0 {
			b.currentPlayer.WinMoney(money)
		}
	}
	return true
}

type SwallowAction struct{}
func (sa SwallowAction)Name()string{
	return "swallow"
}
func (sa SwallowAction)CanMove()bool{return false}
func (sa SwallowAction)Do(b Board,info InfoAction)bool {
	nbCow := b.casesByPosition[info.CasePos].TakeColor(info.Color)
	b.currentPlayer.GetCows(map[int]int8{info.Color:nbCow})

	// if case is empty, remove action
	if !b.casesByPosition[info.CasePos].HasCows() {
		b.casesByPosition[info.CasePos].action = nil
	}
	return true
}
