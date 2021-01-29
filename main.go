package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: verifyme <filepath>")
	}

	fp := os.Args[1]
	selfHash := logVerifierInfo(fp)

	// Make a random cert
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// Marshal the public cert into an Action Output.
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte(pubBytes),
	})

	// base64 encode the entire PEM to fit it in one line.
	outPem := base64.StdEncoding.EncodeToString(pubPem)
	logOutput("publickey", outPem)

	// Now sign and log the input file.
	message, err := ioutil.ReadFile(fp)
	if err != nil {
		log.Fatal(err)
	}
	messageSig, hash := sign(message, priv)
	logOutput("signature", messageSig)
	// Also log the artifact hash
	logOutput("sha256", hash)

	// Pack a metadata envelope and sign that with the same key
	env := envelope{
		RunUrl:      runUrl(),
		GitHubSha:   os.Getenv("GITHUB_SHA"),
		ArtifactSha: hash,
		VerifierSha: selfHash,
	}
	envBytes, err := json.Marshal(env)
	if err != nil {
		log.Fatal(err)
	}
	envStr := base64.StdEncoding.EncodeToString(envBytes)
	envSig, _ := sign([]byte(envStr), priv)

	// Log both the raw and signed environment payloads
	logOutput("environment", envStr)
	logOutput("environment_signature", envSig)
}

func logOutput(key, val string) {
	fmt.Printf("::set-output name=%s::%s\n", key, val)
	fmt.Printf("%s=%s\n", key, val)
}

type envelope struct {
	RunUrl      string
	GitHubSha   string
	ArtifactSha string
	VerifierSha string
	// Add more info here:
	// Roughtime data
	// Actor that did the build
	// Process/system info
}

func sign(b []byte, p *ecdsa.PrivateKey) (string, string) {
	h := sha256.New()
	h.Write(b)
	hash := h.Sum(nil)
	sig, err := ecdsa.SignASN1(rand.Reader, p, hash)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(sig), hex.EncodeToString(hash)
}

func runUrl() string {
	g := os.Getenv
	return fmt.Sprintf("%s/%s/actions/runs/%s", g("GITHUB_SERVER_URL"), g("GITHUB_REPOSITORY"), g("GITHUB_RUN_ID"))
}

func logVerifierInfo(p string) string {
	fmt.Println("Starting verifier with: ", p)

	// Log our own hash
	self, err := os.Open(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer self.Close()
	h := sha256.New()
	if _, err := io.Copy(h, self); err != nil {
		log.Fatal(err)
	}
	hash := hex.EncodeToString(h.Sum(nil))
	fmt.Println("Self hash: ", hash)
	return hash
}
