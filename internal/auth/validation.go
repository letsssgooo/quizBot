package auth

import (
	"errors"
	"regexp"
	"strings"
)

func ParseFullName(message string) (string, error) {
	message = strings.TrimSpace(message)
	if len(strings.Fields(message)) != 3 {
		return "", errors.New("cannot authenticate user, invalid parameters")
	}

	return message, nil
}

func ParseRole(message string) (string, error) {
	message = strings.TrimSpace(message)
	if len(strings.Fields(message)) != 1 {
		return "", errors.New("cannot add role to user, invalid parameter")
	}

	return message, nil
}

func ParseGroup(message string) (string, error) {
	message = strings.TrimSpace(message)
	if len(strings.Fields(message)) != 1 {
		return "", errors.New("cannot add group to user, need only one argument")
	}

	re := regexp.MustCompile(`(?i)^[а-яё]{4}[0-9]{3}$`)
	if !re.MatchString(message) {
		return "", errors.New("cannot add group to user, invalid parameter")
	}

	return message, nil
}
