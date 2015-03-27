package crypt

import (
	"io/ioutil"
	"os"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"bufio"
	"math"
)

// Encrypt file into image ppm


func encryptData(stringKey string,data []byte)[]byte{
	key := []byte(stringKey)

	if rest := len(key)%32 ; len(key) == 0 || rest!=0 {
		key = append(key,make([]byte,32 - rest)...)
	}

	if rest := len(data)%aes.BlockSize ;rest!= 0{
		data = append(data,make([]byte,aes.BlockSize-rest)...)
	}
	cry,_:= aes.NewCipher(key)
	encryptData := make([]byte,len(data))
	encoder := cipher.NewCBCEncrypter(cry,make([]byte,aes.BlockSize))
	encoder.CryptBlocks(encryptData,data)

	return encryptData
}

func decryptData(stringKey string,encodeData []byte)[]byte{
	key := []byte(stringKey)

	if rest := len(key)%32 ; len(key) == 0 || rest!=0 {
		key = append(key,make([]byte,32 - rest)...)
	}

	if rest := len(encodeData)%aes.BlockSize ;rest!= 0{
		encodeData = append(encodeData,make([]byte,aes.BlockSize-rest)...)
	}
	cry,_:= aes.NewCipher(key)

	decoder := cipher.NewCBCDecrypter(cry,make([]byte,aes.BlockSize))
	decodeData := make([]byte,len(encodeData))
	decoder.CryptBlocks(decodeData,encodeData)

	return decodeData
}

func DataToImage(path, source,key string){
	out,_ := os.OpenFile(path,os.O_CREATE|os.O_TRUNC,os.ModePerm)

	in,_ := os.Open(source)
	defer in.Close()
	defer out.Close()


	rawData,_ := ioutil.ReadAll(in)
	// Pad data with aes.BlockSize.
	dataPad := 0
	if dataPad = aes.BlockSize - len(rawData)%aes.BlockSize ; dataPad != 0 {
		rawData = append(rawData,make([]byte,dataPad)...)
	}
	encodeData := encryptData(key,rawData)

	x,y,padTo3 := calcImageSize(int64(len(encodeData)))
	out.Write([]byte(fmt.Sprintf("P6\n%d %d 255 %d %d\n",x,y,dataPad,padTo3)))
	// Pad data to have a 3 multiple of byte (for each pixel)

	if padTo3 > 0{
		out.Write([]byte{byte(padTo3)})
		for i := 0 ; i < padTo3-1 ; i++{
			out.Write([]byte{0})
		}
	}
	// Write data
	out.Write(encodeData)
	fmt.Println("Create image in",path)
}

func ImageToData(source, dest,key string) {
	in,_:= os.Open(source)
	defer in.Close()
	reader := bufio.NewReader(in)

	nb := 2 // Return line to count
	data,_,_ :=reader.ReadLine()
	nb+=len(data)
	data,_,_ =reader.ReadLine()
	nb+=len(data)
	dataPad,padTo3 := 0,0
	fmt.Sscanf(string(data),"%d %d %d %d %d",new(int),new(int),new(int),&dataPad,&padTo3)
	in.Seek(0,0)
	data,_ = ioutil.ReadAll(in)

	zipData := decryptData(key,data[nb+padTo3:])	// Ignore header with nb and pad to 3
	zipData = zipData[:len(zipData)-dataPad]			// Remove pad data at the end of data

	out,_ := os.OpenFile(dest,os.O_CREATE|os.O_TRUNC,os.ModePerm)
	defer out.Close()
	out.Write(zipData)

	fmt.Println("Read image and create file in",dest)
}


func calcImageSize(size int64)(int64,int64,int){
	shift := int(size%3)
	if size%3 != 0 {
		size+=3-size%3
	}
	nbPixels := size/3
	limit := int64(math.Sqrt(float64(size)))

	max := int64(1)
	for i := int64(2) ; i < limit ; i++ {
		if nbPixels % i == 0 {
			// Divisor
			max = i
		}
	}

	return max,nbPixels/max,shift
}



