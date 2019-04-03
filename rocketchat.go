package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/prometheus/alertmanager/template"
)

type Config struct {
	Rocketchat  url.URL
	Credentials models.UserCredentials
}

type RocketChatClient interface {
	GetChannelId(name string) (string, error)
	SendMessage(message *models.Message) (*models.Message, error)
	NewMessage(channel *models.Channel, text string) *models.Message
}

func GetRocketChatAuthenticatedClient(config Config) *realtime.Client {

	rtClient, errClient := realtime.NewClient(&config.Rocketchat, false)
	if errClient != nil {
		log.Printf("Error to get realtime client: %v", errClient)
		return nil
	}

	_, errUser := rtClient.Login(&config.Credentials)
	if errUser != nil {
		log.Printf("Error to login user: %v", errUser)
		return nil
	}

	return rtClient

}

// Need to remove this function once our PR to Rocket.Chat SDK is approved
func newRandomID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%f", rand.Float64())
}

func formatMessage(rtClient RocketChatClient, channel *models.Channel, alert template.Alert) *models.Message {

	const (
		warning       = "warning"
		critical      = "critical"
		warningColor  = "#f2e826"
		criticalColor = "#ef0b1e"
		defaultColor  = "#ffffff"
	)

	var color string
	severity := alert.Labels["severity"]
	title := fmt.Sprintf("**%s: %s**", strings.Title(severity), alert.Annotations["summary"])
	message := rtClient.NewMessage(channel, title)

	if strings.ToLower(severity) == warning {
		color = warningColor
	} else if strings.ToLower(severity) == critical {
		color = criticalColor
	} else {
		color = defaultColor
	}

	var keys []string
	for k := range alert.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	attachementText := fmt.Sprintf("**description**: %s\n", alert.Annotations["description"])
	for _, k := range keys {
		attachementText += fmt.Sprintf("**%s**: %s\n", k, alert.Labels[k])
	}

	message.PostMessage.Attachments = []models.Attachment{
		models.Attachment{
			Color: color,
			Text:  attachementText,
		},
	}
	return message
}

// SendNotification connects to RocketChat server, authenticates the user and sends the notification
func SendNotification(rtClient RocketChatClient, data template.Data) {

	channelID, errRoom := rtClient.GetChannelId(data.CommonLabels["channel_name"])
	if errRoom != nil {
		log.Printf("Error to get room ID: %v", errRoom)
		return
	}
	channel := &models.Channel{ID: channelID}

	log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {

		message := formatMessage(rtClient, channel, alert)
		_, errMessage := rtClient.SendMessage(message)
		if errMessage != nil {
			log.Printf("Error to send message: %v", errMessage)
			return
		}
	}
}
