package msgo

import (
	"strings"
	"unicode"
)

func SubStringLast(url, groupName string) string {
	index := strings.Index(url, groupName)
	if index == -1 {
		return ""
	}

	return url[index+len(groupName):]
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

//func String2Bytes(s string) []byte {
//	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
//	bh := reflect.SliceHeader{
//		Data: stringHeader.Data,
//		Len:  stringHeader.Len,
//	}
//	return *(*[]byte)(unsafe.Pointer(&bh))
//}
