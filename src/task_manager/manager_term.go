package task_manager

import (
    tb "github.com/nsf/termbox-go"
    "fmt"
    "syscall"
    "logger"
    "time"
    "sync"
)

type histo struct {
    tab [][]interface{}
}

type TermLogger struct {
    typeLog string
    logLength int
    logPosition int
    histo * histo
}

// manage available slot to add task
type SlotTasks struct {
    nb int
    // chanel with all availables slots
    slots chan int
}

func NewSlotTasks(length int)*SlotTasks{
    slots := make(chan int,length)
    for i := 0 ; i < length ; i++ {
        slots <- i
    }
    return &SlotTasks{length,slots}
}

// return a number in chanel, if no available, wait a new one ine chanel
func (st * SlotTasks)acquire()int{
    return <- st.slots
}

func (st * SlotTasks)release(id int){
    st.slots <- id
}


func NewTermLogger()(TermLogger,TermLogger) {
    return TermLogger{"INFO : ",4,50,&histo{make([][]interface{},4)}},TermLogger{"ERROR : ",4,50,&histo{make([][]interface{},4)}}
}

func (tl TermLogger)Print(messages ...interface{}){
    tl.histo.tab = append(tl.histo.tab[1:],messages)
    for i := tl.logPosition ; i < tl.logPosition + tl.logLength ; i++ {
        clearLine(i)
    }
    for i,line := range tl.histo.tab{
        if len(line) >0 {
            writeValue(tl.typeLog, 4, tl.logPosition+i)
            pos := 4 + len(tl.typeLog)
            date := time.Now().Format("15:04:05.000")
            writeValue(date, pos, tl.logPosition+i)
            pos+=1 + len(date)
            for _, m := range line {
                message := fmt.Sprintf("%v", m)
                writeValue(message, pos, tl.logPosition+i)
                pos+=len(message)+1
            }
        }
    }
    refresh()
}


type Term struct {
    slotManager * SlotTasks
}

func NewTerm()Term{
    tb.Init()
    //defer tb.Close()

    tb.SetInputMode(tb.InputEsc | tb.InputMouse)
    tb.Clear(tb.ColorDefault, tb.ColorDefault)
    writeValue("TASK MANAGER",20,1)
    refresh()
    t := Term{NewSlotTasks(8)}
    go t.manageEvents()

    return t
}

func (t Term)manageEvents(){
    for {
        event := tb.PollEvent()
        if event.Key == tb.KeyEsc {
            syscall.Exit(1)
        }
    }
}

func (t Term)ShowNodes(nbNodes int){
    writeValue(fmt.Sprintf("%d node(s) connected",nbNodes),50,1)
    refresh()
}

func (t Term)CreateGauge(update chan GaugeData, title string) {
    slot := t.slotManager.acquire()
    g := Gauge{(slot+1) * 5,40}
    g.writeBounds()
    writeValue(title,60,(slot+1) * 5 + 1)
    refresh()
    go func() {
        for {
            gd := <-update
            if gd.total == -1 && gd.current == -1 {
                break
            }
            percent := 0
            if gd.total > 0 {
                percent = (gd.current * 100) / gd.total
            }
            g.writePercent(percent)
            g.writeValue(fmt.Sprintf("%d/%d",gd.current,gd.total))
            refresh()
        }
        close(update)
        logger.GetLogger2().Info("End of batch")
        // Clear gauge
        g.clear()
        t.slotManager.release(slot)
    }()
}

type GaugeData struct {
    current int
    total int
}

type Gauge struct{
    y int
    length int
}

func (g Gauge)clear(){
    for x:= 5 + g.length ; x <=80 ; x++ {
        tb.SetCell(x, g.y+1, ' ', tb.ColorBlack, tb.ColorBlack)
    }
    for line := g.y-1 ; line <=g.y+3 ; line++ {
        for i := 4 ; i < 4 + g.length +1 ; i++ {
            tb.SetCell(i, line, ' ', tb.ColorBlack, tb.ColorBlack)
        }
    }
    refresh()
}

func (g Gauge)writePercent(percent int){
    reelLength := (g.length * percent ) / 100 -2
    for i := 6 ; i < 6 + reelLength ; i++ {
        tb.SetCell(i, g.y+1, ' ', tb.ColorWhite, tb.ColorRed)
    }
    strValue := fmt.Sprintf("%d%%",percent)
    for i := 0 ; i < len(strValue) ; i++ {
        color := tb.ColorBlack
        if (g.length /2) + i <= reelLength {
            color = tb.ColorRed
        }
        tb.SetCell(4 + (g.length /2) + i, g.y+1, rune(strValue[i]), tb.ColorWhite, color)
    }
}

func (g Gauge)writeValue(value string){
    writeValue(value,6+g.length,g.y+1)
}

func (g Gauge) writeBounds(){
    tb.SetCell(4,g.y,'|',tb.ColorWhite,tb.ColorBlack)
    tb.SetCell(4,g.y+1,'|',tb.ColorWhite,tb.ColorBlack)
    tb.SetCell(4,g.y+2,'|',tb.ColorWhite,tb.ColorBlack)

    tb.SetCell(4 + g.length,g.y,'|',tb.ColorWhite,tb.ColorBlack)
    tb.SetCell(4 + g.length,g.y+1,'|',tb.ColorWhite,tb.ColorBlack)
    tb.SetCell(4 + g.length,g.y+2,'|',tb.ColorWhite,tb.ColorBlack)
    for i := 5 ; i < 4+g.length ; i++ {
        tb.SetCell(i, g.y+2, '_', tb.ColorWhite, tb.ColorBlack)
        tb.SetCell(i, g.y-1, '_', tb.ColorWhite, tb.ColorBlack)
    }
    refresh()
}

var lock sync.Mutex= sync.Mutex{}

// avoid 2 refresh at the same time
func refresh(){
    lock.Lock()
    tb.Flush()
    lock.Unlock()
}

func writeValue(value string, x,y int){
    for i := 0 ; i < len(value) ; i++ {
        tb.SetCell(x +i,y,rune(value[i]),tb.ColorWhite,tb.ColorBlack)
    }
}

func clearLine(line int){
    w,_ := tb.Size()
    for i := 0 ; i < w ; i++ {
        tb.SetCell(i, line, ' ', tb.ColorBlack, tb.ColorBlack)
    }
}