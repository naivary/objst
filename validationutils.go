package objst

import (
	"regexp"

	"github.com/google/uuid"
)

func isValidObjectName(name string) bool {
	ok, _ := regexp.MatchString(objectNamePattern, name)
	return ok
}

func isValidUUID(s string) bool {
	if s == "" {
		return true
	}
	_, err := uuid.Parse(s)
	return err == nil
}
