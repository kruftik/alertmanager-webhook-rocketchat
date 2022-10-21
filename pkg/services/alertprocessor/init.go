package alertprocessor

import (
	"html/template"

	"github.com/prometheus/common/log"
)

func init() {
	var err error

	alertBodyTmpl, err = template.New("rc-alert-body").Parse(alertBodyTmplSource)
	if err != nil {
		log.Fatalf("cannot parse alert body template: %v", err)
	}

	alertAttachmentTextTmpl, err = template.New("rc-alert-attachment").Parse(alertAttachmentTextTmplSource)
	if err != nil {
		log.Fatalf("cannot parse alert attachment template: %v", err)
	}
}
