package utils

import "golang.org/x/crypto/bcrypt"

func GeneratePassword(p string) string {
	// Генерируем хеш пароля с использованием bcrypt и дефолтной сложности (bcrypt.DefaultCost).
	hash, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(hash)
}

// ComparePassword сравнивает хешированный пароль с введенным паролем.
func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
