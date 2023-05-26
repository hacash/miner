package minerrelayservice

import "time"

func (r *RelayService) loop() {

	reconnectserver := time.NewTicker(time.Minute * 3)
	clearstuffmaps := time.NewTicker(time.Minute * 5)

	for {
		select {
		case <-reconnectserver.C:
			if r.service_tcp == nil {
				go r.connectToService()
			}
		case <-clearstuffmaps.C:
			if r.penddingBlockStuff != nil {
				var delhei = r.penddingBlockStuff.BlockHeadMeta.GetHeight()
				if delhei > 10 {
					delhei -= 10
					r.prevBlockStuffMaps.Delete(delhei)
				}
			}

		}
	}

}
