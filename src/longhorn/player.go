package longhorn


// Represent of player of longhorn
type Player struct{
	id int
	name string
	// Belong's cows
	cows [4]int8
	//
	moneys []int
}

// @param cows : number of cows organized by color
func (p * Player)GetCows(cows map[int]int8){
	for color,nb := range cows {
		p.cows[color]+=nb
	}
}

func (p * Player)WinMoney(money int){
	p.moneys = append(p.moneys,money)
}

func (p Player)GetTotalMoney()int{
	sum := 0
	for _,money := range p.moneys {
		sum+=money
	}
	return sum
}

// return how many cows are lost
func (p * Player)LoseCows(cows map[int]int8)map[int]int8{
	lostCows := make(map[int]int8,len(cows))
	for color,nb := range cows {
		if nb > p.cows[color] {
			lostCows[color] = p.cows[color]
			p.cows[color] = 0
		}else{
			p.cows[color]-=nb
			lostCows[color] = nb
		}
	}
	return lostCows
}

func (p * Player) StoleCows(p2 * Player,color,nb int8){
	nbColor := p2.cows[color]
	if nbColor >= nb {
		p2.cows[color] -=2
		p.cows[color]+=nb
	}else{
		p2.cows[color] = 0
		p.cows[color]+=nbColor
	}
}

func (p Player)IsWinner()bool{
	// Check if player fot 9 cows of a color
	for _,nb := range p.cows {
		if nb == 9 {
			return true
		}
	}
	return false
}
