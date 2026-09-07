package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    uint32 = 3
	argon2Memory  uint32 = 64 * 1024
	argon2Threads uint8  = 4
	argon2KeyLen  uint32 = 32
	argon2SaltLen        = 16
)

var ErrInvalidPasswordHash = errors.New("invalid password hash")

// HashPassword hashes a password using Argon2id.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	salt := make([]byte, argon2SaltLen)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLen,
	)

	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory,
		argon2Time,
		argon2Threads,
		saltEncoded,
		hashEncoded,
	), nil
}

// VerifyPassword verifies a password against an Argon2id hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	if password == "" {
		return false, nil
	}

	memory, timeCost, threads, salt, expectedHash, err := parseArgon2Hash(encodedHash)
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		timeCost,
		memory,
		threads,
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(actualHash, expectedHash) != 1 {
		return false, nil
	}

	return true, nil
}

func parseArgon2Hash(encodedHash string) (
	memory uint32,
	timeCost uint32,
	threads uint8,
	salt []byte,
	hash []byte,
	err error,
) {
	parts := strings.Split(encodedHash, "$")

	if len(parts) != 6 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	if parts[1] != "argon2id" {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	if parts[2] != "v=19" {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	params := strings.Split(parts[3], ",")

	if len(params) != 3 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	for _, param := range params {
		keyValue := strings.SplitN(param, "=", 2)

		if len(keyValue) != 2 {
			return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
		}

		switch keyValue[0] {
		case "m":
			value, parseErr := strconv.ParseUint(keyValue[1], 10, 32)
			if parseErr != nil {
				return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
			}
			memory = uint32(value)

		case "t":
			value, parseErr := strconv.ParseUint(keyValue[1], 10, 32)
			if parseErr != nil {
				return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
			}
			timeCost = uint32(value)

		case "p":
			value, parseErr := strconv.ParseUint(keyValue[1], 10, 8)
			if parseErr != nil {
				return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
			}
			threads = uint8(value)

		default:
			return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
		}
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	if memory == 0 || timeCost == 0 || threads == 0 || len(salt) == 0 || len(hash) == 0 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	return memory, timeCost, threads, salt, hash, nil
}
