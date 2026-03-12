package helper

import "github.com/sethvargo/go-password/password"

func GenerateRandomPassword(length int) (string, error) {
	passwd, err := password.Generate(length, 4, 2, false, false)
	if err != nil {
		return "", err
	}
	return passwd, nil
}
