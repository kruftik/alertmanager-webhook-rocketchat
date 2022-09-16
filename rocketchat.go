package main

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	amTemplate "github.com/prometheus/alertmanager/template"
	"github.com/prometheus/common/log"
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

	ErrDefaultChannelNotDefined = errors.New("default Rocket.Chat channel name is not defined in configuration")
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

// RocketChat is the client interface to Rocket.Chat
type RocketChat interface {
	Login(credentials *models.UserCredentials) (*models.User, error)
	GetChannelID(channelName string) (string, error)
	SendMessage(message *models.Message) (*models.Message, error)
	NewMessage(channel *models.Channel, text string) *models.Message
}

// RocketChatConnector connector and method base
type RocketChatConnector struct {
	Client *realtime.Client
}

// Login wraps the Login method
func (connector RocketChatConnector) Login(credentials *models.UserCredentials) (*models.User, error) {
	return connector.Client.Login(credentials)
}

// GetChannelID wraps the GetChannelId method
func (connector RocketChatConnector) GetChannelID(channelName string) (string, error) {
	return connector.Client.GetChannelId(channelName)
}

// SendMessage wraps SendMessage method
func (connector RocketChatConnector) SendMessage(message *models.Message) (*models.Message, error) {
	return connector.Client.SendMessage(message)
}

// NewMessage wraps the NewMessage method
func (connector RocketChatConnector) NewMessage(channel *models.Channel, text string) *models.Message {
	return connector.Client.NewMessage(channel, text)
}

// GetRocketChat returns the RocketChat
func GetRocketChat() (RocketChatConnector, error) {
	rtClient, errClient := realtime.NewClient(&config.Endpoint, false)
	if errClient != nil {
		return RocketChatConnector{}, errClient
	}

	return RocketChatConnector{Client: rtClient}, nil

}

// AuthenticateRocketChatClient performs login on the client
func AuthenticateRocketChatClient(connector RocketChat) error {
	_, errUser := connector.Login(&config.Credentials)
	return errUser
}

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

func formatMessage(connector RocketChat, channel *models.Channel, alert amTemplate.Alert, receiver string) (*models.Message, error) {
	var (
		alertBodyBuilder           strings.Builder
		alertAttachmentTextBuilder strings.Builder

		attachmentFields = make([]models.AttachmentField, 0, len(alert.Labels.Names())+len(alert.Annotations.Names()))
	)

	err := alertBodyTmpl.Execute(&alertBodyBuilder, alert)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare alert body: %w", err)
	}

	err = alertAttachmentTextTmpl.Execute(&alertAttachmentTextBuilder, alert)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare alert attachment text: %w", err)
	}

	message := connector.NewMessage(channel, alertBodyBuilder.String())

	attachmentColor := defaultColor

	if alert.Status != resolvedStatus {
		severity := alert.Labels[severityLabel]

		if color, colorExists := config.SeverityColors[severity]; colorExists {
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

// SendNotification connects to RocketChat server, authenticates the user and sends the notification
func SendNotification(connector RocketChat, data amTemplate.Data) error {
	channelName := config.Channel.DefaultChannelName
	if val, ok := data.CommonLabels["channel_name"]; ok {
		channelName = val
	}

	if channelName == "" {
		return ErrDefaultChannelNotDefined
	}

	channelID, errRoom := connector.GetChannelID(channelName)
	if errRoom != nil {
		return fmt.Errorf("cannot get room ID: %w", errRoom)
	}

	channel := &models.Channel{ID: channelID}

	log.Infof("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)

	for _, alert := range data.Alerts {
		message, err := formatMessage(connector, channel, alert, data.Receiver)
		if err != nil {
			return fmt.Errorf("cannot prepare message: %w", err)
		}

		_, errMessage := connector.SendMessage(message)
		if errMessage != nil {
			return fmt.Errorf("cannot send message: %w", errMessage)
		}
	}

	return nil
}
