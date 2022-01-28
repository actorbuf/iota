package gsession

import "github.com/actorbuf/iota/generator/uuid"

func RandomString() string {
	return uuid.TimeUUID().String()
}
