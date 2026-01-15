package engine

import "fmt"

// isCorrectQuiz проверяет на корректность структуру квиза
func isCorrectQuiz(quiz *Quiz) error {
	if quiz.Title == "" {
		return fmt.Errorf("missing field title")
	}

	if quiz.Settings.TimePerQuestion == 0 {
		return fmt.Errorf("missing field time_per_question")
	}

	if quiz.Questions == nil {
		return fmt.Errorf("missing field questions")
	}

	if len(quiz.Questions) == 0 {
		return fmt.Errorf("need at least one question")
	}

	for i, question := range quiz.Questions {
		if question.Text == "" {
			return fmt.Errorf("missing field text of %d question", i)
		}

		if question.Options == nil {
			return fmt.Errorf("missing field options of %d question", i)
		}

		if len(question.Options) < 2 {
			return fmt.Errorf("amount of options must be at least two in %d question", i)
		}

		if question.Correct < 0 {
			return fmt.Errorf("index of correct answer must not be negative in %d question", i)
		}

		if question.Correct >= len(question.Options) {
			return fmt.Errorf("index of correct answer in %d question is out of range", i)
		}
	}

	return nil
}
