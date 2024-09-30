package helper

import "golang.org/x/crypto/bcrypt"

func HashPasswordBcrypt(password string) (hashedPassword string, err error) {
	if password == "" {
		return "", nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
