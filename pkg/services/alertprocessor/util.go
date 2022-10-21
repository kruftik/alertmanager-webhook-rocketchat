package alertprocessor

import (
	"fmt"
	"strings"
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	amTemplate "github.com/prometheus/alertmanager/template"
	"github.com/prometheus/common/log"
)

func formatFields(fields amTemplate.Pairs, hiddenFields map[string]struct{}) ([]models.AttachmentField, error) {
	var (
		attachmentFields = make([]models.AttachmentField, 0, len(fields))
	)

	for _, field := range fields {
		if _, isHiddenField := hiddenFields[field.Name]; isHiddenField {
			continue
		}

		attachmentFields = append(attachmentFields, models.AttachmentField{
			Short: true,
			Title: "**" + field.Name + "**",
			Value: field.Value,
		})
	}

	return attachmentFields, nil
}

func (ap *AlertProcessor) formatMessage(channel *models.Channel, alert amTemplate.Alert, receiver string) (*models.Message, error) {
	var (
		alertBodyBuilder           strings.Builder
		alertAttachmentTextBuilder strings.Builder

		attachmentFields = make([]models.AttachmentField, 0, len(alert.Labels.Names())+len(alert.Annotations.Names()))
	)

	err := ap.tmpl.Body.Execute(&alertBodyBuilder, alert)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare alert body: %w", err)
	}

	err = ap.tmpl.Attachment.Execute(&alertAttachmentTextBuilder, alert)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare alert attachment text: %w", err)
	}

	message := ap.rc.NewMessage(channel, alertBodyBuilder.String())

	attachmentColor := defaultColor

	if alert.Status != resolvedStatus {
		severity := alert.Labels[severityLabel]

		if color, colorExists := ap.cfg.RocketChat.SeverityColors[severity]; colorExists {
			attachmentColor = color
		}
	} else {
		attachmentColor = resolvedColorCode
	}

	labelFields, err := formatFields(alert.Labels.SortedPairs(), hiddenLabels)
	if err != nil {
		return nil, fmt.Errorf("cannot format labels: %w", err)
	}

	attachmentFields = append(attachmentFields, labelFields...)

	annotationFields, err := formatFields(alert.Annotations.SortedPairs(), hiddenAnnotations)
	if err != nil {
		return nil, fmt.Errorf("cannot format annotations: %w", err)
	}

	attachmentFields = append(attachmentFields, annotationFields...)

	message.PostMessage.Attachments = []models.Attachment{
		{
			Color:  attachmentColor,
			Text:   alertAttachmentTextBuilder.String(),
			Fields: attachmentFields,
		},
	}

	return message, nil
}

func retry(retries int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return nil
		}

		if i >= retries {
			break
		}

		time.Sleep(sleep)

		log.Warnf("retrying after error: %v", err)
	}
	return fmt.Errorf("after %d retries, last error: %s", retries, err)
}
