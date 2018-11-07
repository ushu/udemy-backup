package cli

import (
	"github.com/manifoldco/promptui"
	"github.com/ushu/udemy-backup/client"
)

func SelectCourse(courses []client.Course) (*client.Course, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "🤓 {{ .Title | cyan }} ({{ .ID | red }})",
		Inactive: "  {{ .Title | cyan }} ({{ .ID | red }})",
		Selected: "🚀 {{ .Title | red | cyan }}",
		Details: `
--------- Course ----------
{{ "Title:" | faint }}	{{ .Title }}
{{ "Udemy ID:" | faint }}	{{ .ID }}
{{ "URL:" | faint }}	{{ .URL }}`,
	}

	prompt := promptui.Select{
		Label:     "Select course to backup",
		Items:     courses,
		Templates: templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	return &courses[i], nil
}
