package main

import (
	"fmt"
	"log"
)
import "github.com/tarm/serial"

func main() {
	fmt.Println("Bonjour")
	p, err := serial.OpenPort(&serial.Config{Name: "COM7", Baud: 115200})
	if err != nil {
		log.Fatal("Impossible", err)
	}
	log.Println("Starting")
	for {
		data := make([]byte, 256)
		n, err := p.Read(data)
		if err != nil {
			log.Fatal("Error while reading", err)
		}
		if n > 0 {
			log.Println(n, string(data[:n]))
			byteToInt(data[:n])
		}
	}
}

func byteToInt(tab []byte) {
	for _, d := range tab {
		fmt.Print(int(d), " ")
	}
	fmt.Println("")
}
