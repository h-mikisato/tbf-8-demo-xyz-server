package handlers

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"hash"
	"io"
)

func makeInteractionHash(serverNonce, clientNonce, interactionHandle string, hasher hash.Hash) string {
	io.WriteString(hasher, serverNonce)
	io.WriteString(hasher, "\n")
	io.WriteString(hasher, clientNonce)
	io.WriteString(hasher, "\n")
	io.WriteString(hasher, interactionHandle)
	res := hasher.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(res)
}

func getToken() string {
	return getRandomB32(25)
}

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
	for n < len {
		c, _ := rand.Read(seed[n:])
		n += c
	}

	return base32.StdEncoding.EncodeToString(seed)
}
