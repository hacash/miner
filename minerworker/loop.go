package minerworker

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/miner/message"
	"time"
)

// 开始
func (m *MinerWorker) loop() {

	go func() {
		for {
			// 持续不断投喂
			if m.pendingMiningBlockStuff == nil {
				time.Sleep(time.Millisecond)
				continue
			}
			// 开始投喂
			m.miningStuffFeedingCh <- m.pendingMiningBlockStuff.CopyForMiningByRandomSetCoinbaseNonce()
		}
	}()

	for {
		select {

		// 等待挖掘成功
		case result := <-m.miningResultCh:
			var mintSuuessed = result.GetMiningSuccessed()
			if mintSuuessed {
				m.pendingMiningBlockStuff = nil // 重置为空
			}
			if mintSuuessed || m.config.IsReportPower {
				// 上传挖矿结果
				var resupobj = message.MsgReportMiningResult{
					fields.CreateBool(mintSuuessed),
					fields.VarUint5(result.GetHeadMetaBlock().GetHeight()),
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
