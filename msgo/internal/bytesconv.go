package internal

import (
	"reflect"
	"unsafe"
)

func String2Bytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	byteHeader := reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
	}

	return *(*[]byte)(unsafe.Pointer(&byteHeader))
}
