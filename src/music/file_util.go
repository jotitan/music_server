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

func getBytesAsInts32(data []byte)[]int32{
	values := make([]int32,len(data)/4)
	for i := 0 ; i < len(values) ; i++ {
		values[i] = int32(binary.LittleEndian.Uint32(data[i*4:(i+1)*4]))
	}
	return values
}

func getBytesAsInts32Int(data []byte)[]int{
	values := make([]int,len(data)/4)
	for i := 0 ; i < len(values) ; i++ {
		values[i] = int(binary.LittleEndian.Uint32(data[i*4:(i+1)*4]))
	}
	return values
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

func getIntAsByte(n int) []byte {
	return getInt32AsByte(int32(n))
}

func getInts32AsByte(ns []int) []byte {
	tab := make([]byte,0,4*len(ns))
	for _,n := range ns {
		tab = append(tab,getInt32AsByte(int32(n))...)
	}
	return tab
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

func getNextInt32FromFile(f *os.File)int32{
	tab := make([]byte,4)
	f.Read(tab)
	return int32(binary.LittleEndian.Uint32(tab))
}

func getInt64FromBytes(data []byte)int64{
	return int64(binary.LittleEndian.Uint64(data))
}

func getInt32FromBytes(data []byte)int32{
	return int32(binary.LittleEndian.Uint32(data))
}

func getInt32FromFile(f *os.File,pos int64)int32{
	tab := make([]byte,4)
	f.ReadAt(tab,pos)
	return int32(binary.LittleEndian.Uint32(tab))
}

func getInt8FromFile(f *os.File,pos int64)int{
	tab := make([]byte,1)
	f.ReadAt(tab,pos)
	return int(tab[0])
}

func getInt16FromFile(f *os.File,pos int64)int16{
	tab := make([]byte,2)
	f.ReadAt(tab,pos)
	return int16(binary.LittleEndian.Uint16(tab))
}

func writeBytes(to,from []byte,pos int){
	for i := 0 ; i < len(from) ; i++ {
		to[i+pos] = from[i]
	}
}
