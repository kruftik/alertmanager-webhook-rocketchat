package main

import (
	"encoding/json"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"net/url"
	"os"
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

	// Do stuff here
	log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {
		log.Printf("Alert: status=%s,Labels=%v,Annotations=%v", alert.Status, alert.Labels, alert.Annotations)
	}

	// Returns a 200 if everything went smoothly
	sendJSONResponse(w, http.StatusOK, "Success")
}

// Starts 2 listeners
// - first one to give a status on the receiver itself
// - second one to actually process the data
func main() {
	http.HandleFunc("/webhook", webhook)
	http.Handle("/metrics", promhttp.Handler())

	listenAddress := ":9876"
	if os.Getenv("PORT") != "" {
		listenAddress = ":" + os.Getenv("PORT")
	}

	//getClient("prometheus", "prometheus@local.local", "to_be_modified")

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


func getClient(name, email, password string) {

	host := "gossip.dazzlingwrench.fxinnovation.com"
	rtClient, errClient := realtime.NewClient(&url.URL{Scheme: "https", Host: host}, false)
	if errClient != nil {
		log.Printf("Error to get realtime client: %v", errClient)
		return
	}

	credentials := &models.UserCredentials{Name: name, Email: email, Password: password}

	_, errUser := rtClient.Login(credentials)
	if errUser != nil {
		log.Printf("Error to login user: %v", errUser)
		return
	}

	roomID, errRoom := rtClient.GetChannelId("prometheus-test-room")
	if errRoom != nil {
		log.Printf("Error to get room ID: %v", errRoom)
		return
	}
	channel := &models.Channel{ID: roomID}
	_, errMessage := rtClient.SendMessage(channel, "Test AlertManager")
	if errMessage != nil {
		log.Printf("Error to send message: %v", errMessage)
		return
	}

}