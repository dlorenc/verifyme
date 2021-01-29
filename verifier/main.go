package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("USAGE: verifier <signature> <public-key> <message>")
		fmt.Println("artifact and signature should be base64 encoded strings")
		fmt.Println("message should be a filepath or a base64 encoded message")
		os.Exit(1)
	}

	sig, pub, message := os.Args[1], os.Args[2], os.Args[3]
	sigBytes := mustDecode(sig)

	// Parse the public key into something we can use to verify
	pem, _ := pem.Decode(mustDecode(pub))
	if pem.Type != "PUBLIC KEY" {
		log.Panicf("unsupported public key type: %s", pem.Type)
	}
	key, err := x509.ParsePKIXPublicKey(pem.Bytes)
	if err != nil {
		log.Panic(err)
	}
	pubKey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		log.Panicf("unsupported public key format: %s", err)
	}

	// Message can be either a filepath or a string
	hasher := sha256.New()
	f, err := os.Open(message)
	if err == nil {
		defer f.Close()
		if _, err := io.Copy(hasher, f); err != nil {
			log.Panicf("unable to decode or read message: %s", err)
		}
	} else {
		hasher.Write([]byte(message))
	}
	hash := hasher.Sum(nil)

	if ok := ecdsa.VerifyASN1(pubKey, hash, sigBytes); !ok {
		log.Fatal("invalid signature")
	}
	fmt.Println("valid signature")
}

func mustDecode(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Panic(err)
	}
	return b
}
