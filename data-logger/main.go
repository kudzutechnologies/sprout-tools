package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var port = flag.Int("port", 8090, "the port whre to listen for data")
var dataFile = flag.String("file", fmt.Sprintf("data-%s.csv", time.Now().Format("200601021504")), "the .csv file where to log the received data")

func handleData(w http.ResponseWriter, req *http.Request) {
	var p map[string]interface{}
	err := json.NewDecoder(req.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.WithFields(p).Info("Received data sample")

	csv, err := CreateCSVDataFile(*dataFile, ",")
	if err != nil {
		log.Errorf("Could not open log file: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = csv.WriteRecord(p)
	if err != nil {
		log.Errorf("Could not log data: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "OK")
}

func startServer(info *ConnectionInfo) error {
	http.HandleFunc(info.Path, handleData)

	for _, addr := range info.Addresses {
		log.Infof("Listening on: %s:%d on %s path\n", addr, info.Port, info.Path)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", info.Port), nil)
}

func main() {
	flag.Parse()

	info := CreateConnectionInfo(*port, "/data")

	fmt.Println("Kudzu Sprout - Data Logger v1.0")
	fmt.Printf("👉 You can quick-connect your sprout using this link code: %s\n\n", info.String())

	log.Infof("Saving data to: %s\n", *dataFile)
	err := startServer(info)
	if err != nil {
		log.Fatalf("Could not start server: %s", err.Error())
	}
}
