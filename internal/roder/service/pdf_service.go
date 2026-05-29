package service

import (
	"bytes"
	"html/template"
	"os/exec"
	"strings"
)

func GenerateOrderPDF(order any) ([]byte, error) {
	tmpl, err := template.ParseFiles("templates/quote.html")
	if err != nil {
		return nil, err
	}

	var html bytes.Buffer

	err = tmpl.Execute(&html, order)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		"wkhtmltopdf",
		"--page-size", "A4",
		"--dpi", "300",
		"--margin-top", "10",
		"--margin-bottom", "10",
		"--margin-left", "10",
		"--margin-right", "10",
		"--enable-local-file-access",
		"-",
		"-",
	)

	cmd.Stdin = strings.NewReader(html.String())

	var pdf bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &pdf
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return pdf.Bytes(), nil
}
