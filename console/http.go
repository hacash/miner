package console

import (
	"fmt"
	"net/http"
	"strconv"
)

func (mc *MinerConsole) startHttpService() error {

	mux := http.NewServeMux()

	mux.HandleFunc("/", mc.home)

	portstr := strconv.Itoa(mc.config.HttpListenPort)
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
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
