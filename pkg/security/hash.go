package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Hasher struct {
	saltLength uint8
	time       uint32
	memory     uint32
	keyLen     uint32
	threads    uint8
}

func NewHasher(time, memory, KeyLen uint32, threads, saltLength uint8) *Hasher {
	return &Hasher{
		time:       time,
		memory:     memory,
		threads:    threads,
		keyLen:     KeyLen,
		saltLength: saltLength,
	}
}

func (h *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to hash: %v", err)
	}
	hashPassword := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hashPassword)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memory, h.time, h.threads, b64Salt, b64Hash)

	return encodedHash, nil
}

func (h *Hasher) Verify(hashPassword string, password string) (bool, error) {
	var m, t uint32
	var p uint8
	var version int

	parts := strings.Split(hashPassword, "$")

	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash password")
	}

	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("failed to parse version: %w", err)
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false, fmt.Errorf("failed to parse memory, time, threads: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	comparisonHash := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(decodedHash)))

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
