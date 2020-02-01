package handlers

import (
	"crypto/rand"
	"encoding/base32"
)

func getNonce() string {
	return getRandomB32(20)
}

func getHandle() string {
	return getRandomB32(15)
}

func getUserCode() string {
	return getRandomB32(5)
}

func getRandomB32(len int) string {
	var seed = make([]byte, len)
	var n int
	for n != len {
		n, _ = rand.Read(seed)
	}

	return base32.StdEncoding.EncodeToString(seed)
}
