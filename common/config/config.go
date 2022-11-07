package config

import (
	"crypto/ed25519"
	"encoding/hex"
)

type LoadedConfig struct {
	JwtSecretKey         ed25519.PrivateKey
	KafkaBrokers         string
	NewAlbumsTopic       string
	ProcessedAlbumsTopic string
}

var Config *LoadedConfig

func init() {
	seed, err := hex.DecodeString("fafd9d9d26df364baa4889dcadfa29b6243c4eada28ec43683e18da7518d036c")
	if err != nil {
		panic(err)
	}
	Config = &LoadedConfig{
		JwtSecretKey:         ed25519.NewKeyFromSeed(seed),
		KafkaBrokers:         "localhost:29092,localhost:29093,localhost:29094",
		NewAlbumsTopic:       "NewAlbums",
		ProcessedAlbumsTopic: "ProcessedAlbums",
	}
}
