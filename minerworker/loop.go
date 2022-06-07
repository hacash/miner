package minerworker

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"time"
)

// start
func (m *MinerWorker) loop() {

	go func() {
		for {
			// Continuous feeding
			if m.pendingMiningBlockStuff == nil {
				time.Sleep(time.Millisecond * 50)
				continue
			}
			// Start feeding
			m.miningStuffFeedingCh <- m.pendingMiningBlockStuff.CopyForMiningByRandomSetCoinbaseNonce()
		}
	}()

	// Check the connection once every 10 ~ 60 seconds
	checkTcpConnTiker := time.NewTicker(time.Minute * 2)

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
			}
		}
	}

}
