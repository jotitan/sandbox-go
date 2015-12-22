package longhorn

import "errors"


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


	// Switch player or not
	if switchPlayer {
		 b.SwitchPlayer()
	}

	// Send data to user (player info and board data)



	// Check if a winner
	return nil
}
