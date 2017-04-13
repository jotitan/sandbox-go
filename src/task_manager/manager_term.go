package task_manager

import (
    tb "github.com/nsf/termbox-go"
    "fmt"
)

type Term struct {}

func NewTerm()Term{
    tb.Init()
    //defer tb.Close()

    tb.SetInputMode(tb.InputEsc | tb.InputMouse)
    tb.Clear(tb.ColorDefault, tb.ColorDefault)
    return Term{}
}

func (t Term)CreateGauge(update chan GaugeData) {
    g := Gauge{4}
    g.writeBounds(40)
    tb.Flush()
    go func() {
        for {
            gd := <-update
            percent := (gd.current * 100) / gd.total
            g.writePercent(percent, 40)
            tb.Flush()
        }
    }()
}

type GaugeData struct {
    current int
    total int
}

type Gauge struct{
    y int
}

func (g Gauge)writePercent(percent, length int){
    reelLength := (length * percent ) / 100 -2
    for i := 6 ; i < 6 + reelLength ; i++ {
        tb.SetCell(i, g.y+1, ' ', tb.ColorWhite, tb.ColorRed)
    }
    strValue := fmt.Sprintf("%d%%",percent)
    for i := 0 ; i < len(strValue) ; i++ {
        color := tb.ColorBlack
        if (length /2) + i <= reelLength {
            color = tb.ColorRed
        }
        tb.SetCell(4 + (length /2) + i, g.y+1, rune(strValue[i]), tb.ColorWhite, color)
    }
}

func (g Gauge) writeBounds(length int){
    tb.SetCell(4,g.y,'|',tb.ColorWhite,tb.ColorBlack)
    tb.SetCell(4,g.y+1,'|',tb.ColorWhite,tb.ColorBlack)
    tb.SetCell(4,g.y+2,'|',tb.ColorWhite,tb.ColorBlack)

    tb.SetCell(4 + length,g.y,'|',tb.ColorWhite,tb.ColorBlack)
    tb.SetCell(4 + length,g.y+1,'|',tb.ColorWhite,tb.ColorBlack)
    tb.SetCell(4 + length,g.y+2,'|',tb.ColorWhite,tb.ColorBlack)
    for i := 5 ; i < 4+length ; i++ {
        tb.SetCell(i, g.y+2, '_', tb.ColorWhite, tb.ColorBlack)
        tb.SetCell(i, g.y-1, '_', tb.ColorWhite, tb.ColorBlack)
    }
}