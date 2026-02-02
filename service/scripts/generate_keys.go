package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate keys: %v\n", err)
		os.Exit(1)
	}

	// Create keys directory if it doesn't exist
	if err := os.MkdirAll("../keys", 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create keys directory: %v\n", err)
		os.Exit(1)
	}

	// Encode keys as base64
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey)

	// Write private key
	if err := os.WriteFile("../keys/registry_signing_key.pem", []byte(privateKeyBase64), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write private key: %v\n", err)
		os.Exit(1)
	}

	// Write public key
	if err := os.WriteFile("../keys/registry_signing_key.pub", []byte(publicKeyBase64), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Ed25519 key pair generated successfully!")
	fmt.Println("Private key: ../keys/registry_signing_key.pem")
	fmt.Println("Public key: ../keys/registry_signing_key.pub")
	fmt.Println("\nPublic key (for verification):")
	fmt.Println(publicKeyBase64)
}
