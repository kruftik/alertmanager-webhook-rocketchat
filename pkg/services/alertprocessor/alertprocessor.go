package alertprocessor

import (
	"fmt"
	"html/template"
	"time"

	"FXinnovation/alertmanager-webhook-rocketchat/internal/config"
	"FXinnovation/alertmanager-webhook-rocketchat/pkg/services/rocketchat"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	amTemplate "github.com/prometheus/alertmanager/template"
	"github.com/prometheus/common/log"
)

type IAlertProcessor interface {
	SendNotification(data amTemplate.Data) error
}

var (
	retryInterval = 3 * time.Second
	retriesCount  = 3

	_ IAlertProcessor = (*AlertProcessor)(nil)
)

type AlertProcessor struct {
	cfg config.AppConfig
	rc  rocketchat.IRocketChat

	tmpl messageTemplate
}

func New(cfg config.AppConfig, rc rocketchat.IRocketChat) (*AlertProcessor, error) {
	var (
		err error
	)

	if cfg.RocketChat.Channel.DefaultChannelName == "" {
		return nil, ErrDefaultChannelNotDefined
	}

	ap := &AlertProcessor{
		cfg: cfg,
		rc:  rc,
	}

	ap.tmpl, err = ap.parseMessageTemplates()
	if err != nil {
		return nil, fmt.Errorf("cannot parse message templates: %w", err)
	}

	return ap, nil
}

type messageTemplate struct {
	Body       *template.Template
	Attachment *template.Template
}

func (ap *AlertProcessor) parseMessageTemplates() (messageTemplate, error) {
	var (
		tmpl messageTemplate
		err  error
	)

	tmpl.Body, err = template.New("rc-alert-body").Parse(alertBodyTmplSource)
	if err != nil {
		return tmpl, fmt.Errorf("cannot parse alert body template: %w", err)
	}

	tmpl.Attachment, err = template.New("rc-alert-attachment").Parse(alertAttachmentTextTmplSource)
	if err != nil {
		return tmpl, fmt.Errorf("cannot parse alert attachment template: %w", err)
	}

	return tmpl, nil
}

func (ap *AlertProcessor) SendMessageWithRetries(msg *models.Message) error {
	err := retry(retriesCount, retryInterval, func() error {
		_, err := ap.rc.SendMessage(msg)
		if err != nil {
			if _, err := ap.rc.Login(); err != nil {
				return fmt.Errorf("cannot reauthentificate in rocketchat: %w", err)
			}

			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("cannot send message: %w", err)
	}

	return nil
}

// SendNotification connects to RocketChat server, authenticates the user and sends the notification
func (ap *AlertProcessor) SendNotification(data amTemplate.Data) error {
	channelName := ap.cfg.RocketChat.Channel.DefaultChannelName
	if val, ok := data.CommonLabels["channel_name"]; ok {
		channelName = val
	}

	channelID, err := ap.rc.GetChannelID(channelName)
	if err != nil {
		return fmt.Errorf("cannot get room ID: %w", err)
	}

	channel := &models.Channel{
		ID: channelID,
	}

	log.Infof("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)

	for _, alert := range data.Alerts {
		message, err := ap.formatMessage(channel, alert, data.Receiver)
		if err != nil {
			return fmt.Errorf("cannot prepare message: %w", err)
		}

		err = ap.SendMessageWithRetries(message)
		if err != nil {
			return fmt.Errorf("cannot send message: %w", err)
		}
	}

	return nil
}
