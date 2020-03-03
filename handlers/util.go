package handlers

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"hash"
	"io"

	"gopkg.in/square/go-jose.v2"
)

func compareKey(this, that jose.JSONWebKey) bool {
	var (
		thisP []byte
		thatP []byte
		err   error
	)
	thisP, err = this.Thumbprint(crypto.SHA256)
	if err != nil {
		return false
	}
	thatP, err = that.Thumbprint(crypto.SHA256)
	if err != nil {
		return false
	}
	return bytes.Compare(thisP, thatP) == 0
}

func makeInteractionHash(serverNonce, clientNonce, interactionHandle string, hasher hash.Hash) string {
	io.WriteString(hasher, serverNonce)
	io.WriteString(hasher, "\n")
	io.WriteString(hasher, clientNonce)
	io.WriteString(hasher, "\n")
	io.WriteString(hasher, interactionHandle)
	res := hasher.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(res)
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
