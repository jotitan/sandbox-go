package main
import (
    "fmt"
    "os"
    "bufio"
	"encoding/hex"
    "encoding/binary"
)



func main(){

    path := "src/resources/img.jpg"
	f,_ := os.Open(path)

    defer f.Close()
    buf := bufio.NewReader(f)
    line,_,_ := buf.ReadLine()

    fmt.Println(len(line),line)
    fmt.Println(string(line),"\n\n")
    // Read Start marker
    pos := 0
    for {
        start := hex.EncodeToString(line[pos:pos+2])
        switch start {
            case "ffd8" : fmt.Println("Start")
            pos+=2
            case "ffe0" : fmt.Println("App marker")
                size := int(binary.BigEndian.Uint16(line[pos+2:pos+4]))
                data := line[pos+4:pos+2+size]
                fmt.Println(data,string(data))
                pos+=2+size
            case "ffe1" : fmt.Println("Exif")
                size := int(binary.BigEndian.Uint16(line[pos+2:pos+4]))
                exifHeader := string(line[pos+4:pos+10])
                bitIndien := string(line[pos+10:pos+12])
                rest := string(line[pos+12:pos+18])
                fmt.Println(size,exifHeader,bitIndien,rest)
                nbMarker := int(binary.LittleEndian.Uint16(line[pos+18:pos+20]))
                pos+=20
                for i :=0 ; i < nbMarker ; i++{
                    fmt.Println(line[pos:pos+12])
                    pos+=12
                }
                return
            default:return
        }

    }
    return

    // Application marker

    // Read ifd0
    return
    // Read exif header
    // Read tiff header

    // Read nb marker

    // Read markers
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
