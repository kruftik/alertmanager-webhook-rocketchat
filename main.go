package main

import (
	"encoding/json"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"net/http"
	"os"
)

var (
	configFile = kingpin.Flag("config.file", "RocketChat configuration file.").Default("config/rocketchat.yml").String()
)

// Webhook http response
type JSONResponse struct {
	Status  int
	Message string
}

func webhook(w http.ResponseWriter, r *http.Request) {

	// Extract data from the body in the Data template provided by AlertManager
	data := template.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Do not forget to close the body at the end
	defer r.Body.Close()

	rocketChatClient := GetRocketChatClient(*configFile)

	// Format notifications and send it
	SendNotification(rocketChatClient, data)

	// Returns a 200 if everything went smoothly
	sendJSONResponse(w, http.StatusOK, "Success")
}

// Starts 2 listeners
// - first one to give a status on the receiver itself
// - second one to actually process the data
func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	http.HandleFunc("/webhook", webhook)
	http.Handle("/metrics", promhttp.Handler())

	listenAddress := ":9876"
	if os.Getenv("PORT") != "" {
		listenAddress = ":" + os.Getenv("PORT")
	}

	log.Printf("listening on: %v", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func sendJSONResponse(w http.ResponseWriter, status int, message string) {
	data := JSONResponse{
		Status:  status,
		Message: message,
	}
	bytes, _ := json.Marshal(data)

	w.WriteHeader(status)
	w.Write(bytes)
}

