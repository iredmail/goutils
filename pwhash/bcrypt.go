package pwhash

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func GenerateBcryptPassword(password string) (hash string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	hash = "{CRYPT}" + string(hashedPassword)

	return
}

func VerifyBcryptPassword(challengePassword, plainPassword string) bool {
	if strings.HasPrefix(challengePassword, "{CRYPT}$2a$") ||
		strings.HasPrefix(challengePassword, "{CRYPT}$2b$") ||
		strings.HasPrefix(challengePassword, "{crypt}$2a$") ||
		strings.HasPrefix(challengePassword, "{crypt}$2b$") {
		challengePassword = challengePassword[7:]
	} else if strings.HasPrefix(challengePassword, "{BLF-CRYPT}") ||
		strings.HasPrefix(challengePassword, "{blf-crypt}") {
		challengePassword = challengePassword[11:]
	}

	return bcrypt.CompareHashAndPassword([]byte(challengePassword), []byte(plainPassword)) == nil
}
