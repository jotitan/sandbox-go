package longhorn


// manage message between server and client

type ServerMessage struct{
	// info about player status, game status
	NextPlayer int
	// case where player are
	CurrentCase int8
	// action to play (if only one color stayed on case)
	Moves map[int][]int
	// info about player
	P1 Player
	P2 Player
}

func NewServerMessage(b Board)ServerMessage{
	// cases reachable for each color (based on number of cow)
	moves := make(map[int][]int)
	if b.currentCase != nil {
		for color, nb := range b.currentCase.GetCowsByColor() {
			moves[color] = b.FindPlayableCases(int(b.currentCase.position), nb)
		}
	}
	casePosition := int8(-1)
	if b.currentCase!= nil {
		casePosition = b.currentCase.position
	}
	return ServerMessage{b.currentPlayer.id,casePosition,moves,b.p1,b.p2}
}
