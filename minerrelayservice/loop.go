package minerrelayservice

import "time"

func (r *RelayService) loop() {

	reconnectserver := time.NewTicker(time.Minute * 3) // 两分钟检查一次重连

	for {
		select {
		case <-reconnectserver.C:
			if r.service_tcp == nil {
				go r.connectToService()
			}

		}
	}

}
