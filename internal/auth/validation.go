package auth

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ParseStudentsData валидирует сообщение студента и отдает обработанные данные (фио и группа)
func ParseStudentsData(data []string) ([]string, error) {
	fullNameParts := data[:len(data)-1]

	if len(fullNameParts) != 3 {
		return nil, fmt.Errorf("%w, need 3 parts of fullname", ErrValidation)
	}

	for i, word := range fullNameParts {
		wordRunes := []rune(word)
		if len(wordRunes) < 2 {
			return nil, fmt.Errorf("%w, there are too few letters in a part of fullName", ErrValidation)
		}

		wordRunes[0] = unicode.ToUpper(wordRunes[0])

		for j := 1; j < len(wordRunes); j++ {
			letter := wordRunes[j]
			if letter == '-' {
				continue
			}

			if !unicode.IsLetter(letter) {
				return nil, fmt.Errorf("%w, only letters and '-' can be in fullName", ErrValidation)
			}

			wordRunes[j] = unicode.ToLower(wordRunes[j])
		}

		fullNameParts[i] = string(wordRunes)
	}

	group := data[len(data)-1]

	re := regexp.MustCompile(`(?i)^[а-яё]{4}[0-9]{3}$`)
	if !re.MatchString(group) {
		return nil, fmt.Errorf("%w, cannot add group to user, invalid parameter", ErrValidation)
	}

	group = strings.ToUpper(group)

	return []string{strings.Join(fullNameParts, " "), group}, nil
}

// ParseRole валидирует сообщение пользователя и отдает роль в виде строки
func ParseRole(message string) (string, error) {
	message = strings.TrimSpace(message)
	if len(strings.Fields(message)) != 1 {
		return "", fmt.Errorf("%w, cannot add role to user, invalid parameterr", ErrValidation)
	}

	return strings.ToLower(message), nil
}
