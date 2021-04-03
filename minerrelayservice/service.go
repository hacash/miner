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
		// 不启动服务器
		fmt.Println("config http_api_listen_port==0 do not start http api service.")
		return

	}

	api.initRoutes()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ResponseData(w, ResponseCreateData("service", "hacash miner relay service"))
	})

	// 路由
	mux.HandleFunc("/query", api.dealQuery)         // 查询
	mux.HandleFunc("/create", api.dealCreate)       // 创建
	mux.HandleFunc("/submit", api.dealSubmit)       // 提交
	mux.HandleFunc("/operate", api.dealOperate)     // 修改
	mux.HandleFunc("/calculate", api.dealCalculate) // 计算

	// 设置监听的端口
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
