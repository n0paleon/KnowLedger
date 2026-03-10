package utils

import (
	"github.com/oklog/ulid/v2"
)

func GenerateRandomULID() string {
	return ulid.Make().String()
}
