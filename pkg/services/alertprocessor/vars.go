package alertprocessor

import (
	"html/template"
)

const (
	alertBodyTmplSource = `**{{ .Labels.alertname }}**: [**{{ .Labels.severity }}**] {{ .Annotations.summary }} {{ if .Annotations.message }}| {{ .Annotations.message }}{{ end }}`

	alertAttachmentTextTmplSource = `{{ .Annotations.description }}`

	resolvedStatus = "resolved"

	alertNameLabel = "alertname"
	severityLabel  = "severity"

	resolvedColorCode = "#00994c"

	defaultColor = "#ffffff"
)

var (
	hiddenLabels = map[string]struct{}{
		alertNameLabel: {},
		severityLabel:  {},
	}

	hiddenAnnotations = map[string]struct{}{
		"description": {},
		"message":     {},
		"summary":     {},
		//"runbook_url": {},
	}

	alertBodyTmpl           *template.Template
	alertAttachmentTextTmpl *template.Template
)
