package main
import (
    "fmt"
    "os"
    "bufio"
)



func main(){

    path := "D:\\temp\\resize\\IMG_2473_DxO.jpg"

    f,_ := os.Open(path)
    defer f.Close()
    buf := bufio.NewReader(f)
    line,_,_ := buf.ReadLine()
    pos := 2

    fmt.Println(len(line),line)
    fmt.Println(string(line))
    for  {
        if line[pos]!=255 {
            break
        }
        marker := line[pos+1]
        length := int(line[pos+2])  + int(line[pos+3])
        fmt.Println("PRE",marker,length,pos)
        data := line[pos+4:pos+2+length]
        pos+=2+length
        fmt.Println("MARKER",marker,pos,string(data))
        fmt.Println("REST",string(line[pos:]))

    }
}
