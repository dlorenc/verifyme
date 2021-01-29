package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	filepath := os.Args[1]
	// Make a random cert
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	pub := priv.PublicKey

	pubBytes, err := x509.MarshalPKIXPublicKey(&pub)
	if err != nil {
		log.Fatal(err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte(pubBytes),
	})

	// base64 encode it one more time
	outPem := base64.StdEncoding.EncodeToString(pubPem)

	fmt.Printf("::set-output name=publickey::%s\n", outPem)

	// Now sign the file.
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	h := sha256.New()
	h.Write(b)
	hash := h.Sum(nil)
	sig1Bytes, err := ecdsa.SignASN1(rand.Reader, priv, hash)
	if err != nil {
		log.Fatal(err)
	}
	sig1 := base64.StdEncoding.EncodeToString(sig1Bytes)

	fmt.Printf("::set-output name=signature::%s\n", sig1)
}
