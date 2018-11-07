package cli

import (
	"errors"

	"github.com/spf13/viper"

	"github.com/manifoldco/promptui"
)

func EnsureCredentials() (string, string, error) {
	id := viper.GetString("id")
	token := viper.GetString("token")

	if id != "" && token != "" {
		return id, token, nil
	}

	return AskCredentials()
}

func AskCredentials() (string, string, error) {
	prompt := promptui.Prompt{
		Label:    "Udemy ID",
		Validate: notEmpty,
	}
	id, err := prompt.Run()
	if err != nil {
		return "", "", err
	}

	prompt = promptui.Prompt{
		Label:    "Udemy Token",
		Validate: notEmpty,
	}
	token, err := prompt.Run()
	if err != nil {
		return "", "", err
	}

	viper.Set("id", id)
	viper.Set("token", token)
	return id, token, nil
}

func notEmpty(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}
	return nil
}
