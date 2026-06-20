package email

import (
	"bytes"
	"fmt"
	"text/template"
)

func LoadTemplates(filepath string, data any) (string, error) {
	tmpl, err := template.ParseFiles(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var htmlBody bytes.Buffer
	err = tmpl.Execute(&htmlBody, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return htmlBody.String(), nil
}
