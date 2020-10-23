package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func startupMessages(staticRoot string, port string) (message string) {
	message = "Demo webserver\n"
	message += fmt.Sprintf("Serving content from %s on port: %s\n", staticRoot, port)

	log.Printf("Serving content from %s on port: %s\n", staticRoot, port)

	return message
}

func main() {
	var err error
	staticRoot := "/app/static"
	if localRun := (os.Getenv("RUNNING_LOCAL") == "YES"); localRun {
		staticRoot, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		staticRoot = staticRoot + "/static"
		log.Println("Running locally...")
	}
	port := "8080"

	fileServer := http.FileServer(http.Dir(staticRoot))
	http.Handle("/", fileServer)

	log.Printf(startupMessages(staticRoot, port))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
