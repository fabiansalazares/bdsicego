package utils

import (
	"bytes"
)

func ReplaceSpanishCharacters(inString string) string {
	var outString string

	outString = string(bytes.Replace([]byte(inString), []byte{0xc1}, []byte{0x41}, -1))  // Á -> A
	outString = string(bytes.Replace([]byte(outString), []byte{0xc9}, []byte{0x45}, -1)) // É -> E
	outString = string(bytes.Replace([]byte(outString), []byte{0xcd}, []byte{0x49}, -1)) // Í -> I
	outString = string(bytes.Replace([]byte(outString), []byte{0xd3}, []byte{0x4f}, -1)) // Ó -> O
	outString = string(bytes.Replace([]byte(outString), []byte{0xda}, []byte{0x55}, -1)) // Ú -> U
	outString = string(bytes.Replace([]byte(outString), []byte{0xd1}, []byte{0x4e}, -1)) // Ñ -> N

	return outString
}
