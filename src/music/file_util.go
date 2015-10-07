package music

import (
	"os"
	"encoding/binary"
)



func getInts64AsByte(values []int64) []byte {
	tab := make([]byte,len(values)*8)
	for i,n := range values {
		writeBytes(tab,getInt64AsByte(n),i*8)
	}
	return tab
}

// GetInt64AsByte return a byte array representation of int64
func getInt64AsByte(n int64) []byte {
	return []byte{byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24),
		byte(n >> 32), byte(n >> 40), byte(n >> 48), byte(n >> 56),
	}
}

// GetInt32AsByte return a byte array representation of int32
func getInt32AsByte(n int32) []byte {
	return []byte{byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24)}
}


// GetInt16AsByte return a byte array representation of int16
func getInt16AsByte(n int16) []byte {
	return []byte{byte(n), byte(n >> 8)}
}

func getInt64FromFile(f *os.File,pos int64)int64{
	tab := make([]byte,8)
	f.ReadAt(tab,pos)
	return int64(binary.LittleEndian.Uint64(tab))
}

func getInt32FromFile(f *os.File,pos int64)int32{
	tab := make([]byte,4)
	f.ReadAt(tab,pos)
	return int32(binary.LittleEndian.Uint32(tab))
}

func writeBytes(to,from []byte,pos int){
	for i := 0 ; i < len(from) ; i++ {
		to[i+pos] = from[i]
	}
}
