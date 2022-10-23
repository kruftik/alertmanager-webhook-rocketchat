package rocketchat

import (
	"fmt"
	"log"

	"FXinnovation/alertmanager-webhook-rocketchat/internal/config"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
)

// IRocketChat is the client interface to Rocket.Chat
type IRocketChat interface {
	Login() (*models.User, error)

	GetChannelID(channelName string) (string, error)
	SendMessage(message *models.Message) (*models.Message, error)
	NewMessage(channel *models.Channel, text string) *models.Message

	CheckAuthSessionStatus() error
}

var (
	_ IRocketChat = (*Service)(nil)
)

// Service provides base functionality via rocketchat API
type Service struct {
	Client *realtime.Client

	Credentials *models.UserCredentials
}

func New(cfg config.AppConfig) (*Service, error) {
	rtClient, err := realtime.NewClient(&cfg.RocketChat.Endpoint, false)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize RC client: %w", err)
	}

	rc := &Service{
		Client:      rtClient,
		Credentials: &cfg.RocketChat.Credentials,
	}

	if _, err := rc.Login(); err != nil {
		return nil, fmt.Errorf("cannot authorize in rocketchat: %w", err)
	}

	return rc, nil
}

// Login wraps the Login method
func (svc Service) Login() (*models.User, error) {
	user, err := svc.Client.Login(svc.Credentials)
	if err != nil {
		log.Printf("rocketchat login error:  >>> %#v <<<", err)

		return nil, err
	}

	return user, nil
}

// GetChannelID wraps the GetChannelId method
func (svc Service) GetChannelID(channelName string) (string, error) {
	return svc.Client.GetChannelId(channelName)
}

// SendMessage wraps SendMessage method
func (svc Service) SendMessage(message *models.Message) (*models.Message, error) {
	return svc.Client.SendMessage(message)
}

// NewMessage wraps the NewMessage method
func (svc Service) NewMessage(channel *models.Channel, text string) *models.Message {
	return svc.Client.NewMessage(channel, text)
}

func (svc Service) CheckAuthSessionStatus() error {
	_, err := svc.Client.GetChannelsIn()
	if err != nil {
		return ErrAuthSessionExpired
	}

	return nil
}
