package minerrelayservice

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (api *RelayService) startHttpApiService() {

	port := api.config.HttpApiListenPort
	if port == 0 {
		// Do not start the server
		fmt.Println("config http_api_listen_port==0 do not start http api service.")
		return

	}

	api.initRoutes()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ResponseData(w, ResponseCreateData("service", "hacash miner relay service"))
	})

	// route
	mux.HandleFunc("/query", api.dealQuery)         // query
	mux.HandleFunc("/create", api.dealCreate)       // establish
	mux.HandleFunc("/submit", api.dealSubmit)       // Submit
	mux.HandleFunc("/operate", api.dealOperate)     // modify
	mux.HandleFunc("/calculate", api.dealCalculate) // calculation

	// Set listening port
	portstr := strconv.Itoa(port)
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[Miner Relay Service] Http api listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}
