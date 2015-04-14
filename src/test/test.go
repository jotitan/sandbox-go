package main
import (
    "fmt"
    "os"
    "bufio"
	"encoding/hex"
)



func main(){

    path := "src/resources/img.jpg"
	f,_ := os.Open(path)

    defer f.Close()
    buf := bufio.NewReader(f)
    line,_,_ := buf.ReadLine()
    pos := 2

	fmt.Println([]byte("Canon"))

    fmt.Println(len(line),line)
    fmt.Println(string(line),"\n\n")
    for  {
        if line[pos]!=255 {
            break
        }
		marker := hex.EncodeToString(line[pos:pos+2])
		switch marker {
			case "ffe0" : fmt.Println("INF : App marker")
			case "ffe1" : fmt.Println("INF : Tiff header")
		}
        length := int(line[pos+2])  + int(line[pos+3])
        fmt.Println("PRE",length,pos)
        data := line[pos+4:pos+2+length]
        pos+=2+length
        fmt.Println("MARKER ",pos,string(data))
    }
	line,_,_ = buf.ReadLine()
	line,_,_ = buf.ReadLine()
	line,_,_ = buf.ReadLine()
	fmt.Println("==>",string(line),"\n\n")
}
