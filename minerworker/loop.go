package minerworker

import (
	"time"
)

// start
func (m *MinerWorker) loop() {

	// Check the connection once every 10 ~ 60 seconds
	checkTcpConnTiker := time.NewTicker(time.Minute * 6)

	for {
		select {

		// Check connection
		case <-checkTcpConnTiker.C:
			//fmt.Println("<-checkTcpConnTiker.C:", m.conn)
			if m.conn == nil {
				//fmt.Println("go startConnect()")
				// Initiate reconnection
				go m.startConnect()
			}

			/*
				// Wait for successful mining
				case result := <-m.miningResultCh:
					var mintSuccessed = result.GetMiningSuccessed()
					if mintSuccessed {
						m.pendingMiningBlockStuff = nil // Reset to null
					}
					if mintSuccessed || m.config.IsReportHashrate {
						// Upload mining results
						var resupobj = message.MsgReportMiningResult{
							fields.CreateBool(mintSuccessed),
							fields.BlockHeight(result.GetHeadMetaBlock().GetHeight()),
							result.GetHeadNonce(),
							result.GetCoinbaseNonce(),
						}
						resbts := resupobj.Serialize()
						// tcp send
						if m.conn != nil {
							message.MsgSendToTcpConn(m.conn, message.MinerWorkMsgTypeReportMiningResult, resbts)
						}
					}*/
		}
	}

}
