package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/prometheus/alertmanager/template"
	"gopkg.in/yaml.v2"
)

// Config struct for Rocketchat credentials and url
type Config struct {
	Rocketchat  url.URL
	Credentials models.UserCredentials
}

// GetRocketChatClient takes path to config file and returns *realtime.Client
func GetRocketChatClient(configFile string) *realtime.Client {

	config := Config{}

	// Load the config from the file
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	errYAML := yaml.Unmarshal([]byte(configData), &config)
	if errYAML != nil {
		log.Fatalf("Error: %v", errYAML)
	}

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

func formatMessage(channel *models.Channel, alert template.Alert) *models.Message {

	const (
		warning       = "warning"
		critical      = "critical"
		warningColor  = "#f2e826"
		criticalColor = "#ef0b1e"
		defaultColor  = "#ffffff"
	)

	var color string
	title := fmt.Sprintf("**%s: %s**", strings.Title(alert.Status), alert.Annotations["summary"])
	attachementText := fmt.Sprintf("**description**: %s\n", alert.Annotations["description"])

	if strings.ToLower(alert.Status) == warning {
		color = warningColor
	} else if strings.ToLower(alert.Status) == critical {
		color = criticalColor
	} else {
		color = defaultColor
	}

	for k, v := range alert.Labels {
		attachementText += fmt.Sprintf("**%s**: %s\n", k, v)
	}

	return &models.Message{
		ID:     newRandomID(),
		RoomID: channel.ID,
		Msg:    title,
		PostMessage: models.PostMessage{
			Attachments: []models.Attachment{
				models.Attachment{
					Color: color,
					Text:  attachementText,
				},
			},
		},
	}
}

// SendNotification connects to RocketChat server, authenticate the user and send the notification
func SendNotification(rtClient *realtime.Client, data template.Data) {

	channelID, errRoom := rtClient.GetChannelId(data.CommonLabels["channel_name"])
	if errRoom != nil {
		log.Printf("Error to get room ID: %v", errRoom)
		return
	}
	channel := &models.Channel{ID: channelID}

	log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {

		message := formatMessage(channel, alert)

		_, errMessage := rtClient.SendMessage(message)
		if errMessage != nil {
			log.Printf("Error to send message: %v", errMessage)
			return
		}
	}
}
