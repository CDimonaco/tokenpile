package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
)

func EnsureIdentity(p Paths) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	if _, err := os.Stat(p.IdentityKeyPath); err == nil {
		return loadIdentity(p)
	}

	return generateIdentity(p)
}

func generateIdentity(p Paths) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ed25519 key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: priv.Seed(),
	})

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "ED25519 PUBLIC KEY",
		Bytes: pub,
	})

	if err = os.WriteFile(p.IdentityKeyPath, privPEM, 0o600); err != nil {
		return nil, nil, fmt.Errorf("write identity key: %w", err)
	}

	if err = os.WriteFile(p.IdentityPubPath, pubPEM, 0o644); err != nil { //nolint:gosec
		return nil, nil, fmt.Errorf("write identity pub: %w", err)
	}

	slog.Info("generated signing identity", "path", p.IdentityKeyPath)

	return priv, pub, nil
}

func loadIdentity(p Paths) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	data, err := os.ReadFile(p.IdentityKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read identity key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, nil, fmt.Errorf("decode identity key PEM")
	}

	if len(block.Bytes) != ed25519.SeedSize {
		return nil, nil, fmt.Errorf("invalid identity key size: got %d, want %d", len(block.Bytes), ed25519.SeedSize)
	}

	priv := ed25519.NewKeyFromSeed(block.Bytes)

	return priv, priv.Public().(ed25519.PublicKey), nil
}
