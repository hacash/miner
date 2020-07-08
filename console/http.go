package console

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (mc *MinerConsole) startHttpService() error {

	mux := http.NewServeMux()

	mux.HandleFunc("/", mc.home)
	mux.HandleFunc("/api/console", mc.console)
	mux.HandleFunc("/api/addresses", mc.addresses)

	portstr := strconv.Itoa(mc.config.HttpListenPort)
	server := &http.Server{
		ReadTimeout:       time.Second * 10,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 10,
		IdleTimeout:       time.Second * 10,
		Addr:              ":" + portstr,
		Handler:           mux,
	}

	fmt.Println("[Miner Pool Console] Http service listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

func parseRequestQuery(request *http.Request) map[string]string {
	request.ParseForm()
	params := make(map[string]string, 0)
	for k, v := range request.Form {
		//fmt.Println("key:", k)
		//fmt.Println("val:", strings.Join(v, ""))
		params[k] = strings.Join(v, "")
	}
	return params
}
