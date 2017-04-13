package longhorn


// Represent of player of longhorn
type Player struct{
	Id int
	Name string
	// Belong's cows
	Cows [4]int8
	//
	Moneys []int
	// Point of the player
	Point int
}

// @param cows : number of cows organized by color
func (p * Player)GetCows(cows map[int]int8){
	for color,nb := range cows {
		p.Cows[color]+=nb
	}
}

func (p * Player)Reset(){
	p.Moneys = make([]int,0)
	p.Cows = [4]int8{}
}

func (p * Player)WinMoney(money int){
	p.Moneys = append(p.Moneys,money)
}

func (p Player)GetTotalMoney()int{
	sum := 0
	for _,money := range p.Moneys {
		sum+=money
	}
	return sum
}

// LoseMoney remove hightest money and return value
func (p * Player)LoseMoney()int{
	if len(p.Moneys) == 0 {
		return 0
	}
	max := p.Moneys[0]
	pos := 0
	for i,money := range p.Moneys {
		if money > max {
			max = money
			pos = i
		}
	}
	p.Moneys = append(p.Moneys[:pos],p.Moneys[pos+1:]...)
	return max
}

// return how many cows are lost
func (p * Player)LoseCows(cows map[int]int8)map[int]int8{
	lostCows := make(map[int]int8,len(cows))
	for color,nb := range cows {
		if nb > p.Cows[color] {
			lostCows[color] = p.Cows[color]
			p.Cows[color] = 0
		}else{
			p.Cows[color]-=nb
			lostCows[color] = nb
		}
	}
	return lostCows
}

// boardCows contains nb of cows on board by color
func (p Player)CountPoint(boardCows []int)int{
	// Point if 100 for each color staying on board
	total := 0
	for i,nb := range p.Cows {
		total+=boardCows[i]*100 * int(nb)
	}
	for _,money := range p.Moneys {
		total+=money
	}
	return total
}

func (p * Player) StoleCows(p2 * Player,color,nb int8){
	nbColor := p2.Cows[color]
	if nbColor >= nb {
		p2.Cows[color] -=2
		p.Cows[color]+=nb
	}else{
		p2.Cows[color] = 0
		p.Cows[color]+=nbColor
	}
}

func (p Player)IsWinner()bool{
	// Check if player fot 9 cows of a color
	for _,nb := range p.Cows {
		if nb == 9 {
			return true
		}
	}
	return false
}
