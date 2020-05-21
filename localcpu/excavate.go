package localcpu

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"sync"
)

// do mining
func (l *LocalCPUPowMaster) Excavate(inputblockheadmeta interfaces.Block, outputCh chan message.PowMasterMsg) {

	//fmt.Println(" --------------------  (l *LocalCPUPowMaster) Excavate")

	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	l.StopMining() // stop old all

	maxuint32 := uint32(4294967295)
	supervene := uint32(l.config.Concurrent) // 并发线程
	// -------- test start --------
	//maxuint32 = uint32(2294900)
	//supervene := uint32(6)
	// -------- test end   --------
	nonceSpace := maxuint32 / supervene

	// new run
	var nextstop byte = 0
	l.stopMarks.Store(&nextstop, &nextstop)

	//fmt.Println("	go func(stopmark *byte) {   ")

	go func(stopmark *byte) {

		defer func() {
			//fmt.Println("	l.stopMarks.Delete(&nextstop) // clean  ")
			l.stopMarks.Delete(&nextstop) // clean
		}()

		var successMiningMark uint32 = 0
		var successFindBlock bool = false

		var miningBlockCh = make(chan miningBlockReturn, 1)
		if l.config.ReturnPowerHash {
			miningBlockCh = make(chan miningBlockReturn, supervene)
		}
		var syncWait = sync.WaitGroup{}
		syncWait.Add(int(supervene))

		for i := uint32(0); i < supervene; i++ {
			//fmt.Println("worker := NewCPUWorker ", i)
			worker := NewCPUWorker(&successMiningMark, miningBlockCh, 0, stopmark)
			if l.config.ReturnPowerHash {
				worker.returnPowerHash = true
			}
			worker.coinbaseMsgNum = l.coinbaseMsgNum
			//l.currentWorkers.Add( worker )
			go func(startNonce, endNonce uint32) {
				//fmt.Println( "start worker.RunMining" )
				success := worker.RunMining(inputblockheadmeta, startNonce, endNonce)
				//fmt.Println( "end worker.RunMining", success )
				if success {
					successFindBlock = true
				}
				syncWait.Done()
				//fmt.Println( "end syncWait.Done()" )
			}(nonceSpace*i, nonceSpace*i+nonceSpace)
		}

		//fmt.Println("syncWait.Wait()  start ", supervene)
		// wait
		syncWait.Wait()

		//fmt.Println("syncWait.Wait()  end ")

		//fmt.Println(successFindBlock)

		if successFindBlock == true {

			// success
			success := <-miningBlockCh
			outputCh <- message.PowMasterMsg{
				Status:         message.PowMasterMsgStatusSuccess,
				CoinbaseMsgNum: fields.VarInt4(success.coinbaseMsgNum),
				NonceBytes:     success.nonceBytes,
				BlockHeadMeta:  inputblockheadmeta,
			}
			return // ok end

		}

		if l.config.ReturnPowerHash == false {

			if *stopmark == 1 {
				// stop
				outputCh <- message.PowMasterMsg{
					Status: message.PowMasterMsgStatusStop,
				}
				return
			}

			if successFindBlock == false {
				// continue
				outputCh <- message.PowMasterMsg{
					Status:         message.PowMasterMsgStatusContinue,
					CoinbaseMsgNum: fields.VarInt4(l.coinbaseMsgNum),
					BlockHeadMeta:  inputblockheadmeta,
				}
				return
			}

		} else {

			// 统计并对比挖矿哈希结果
			var mostPowerHashNonceBytes []byte = nil
			var mostPowerHash []byte = nil
			var mostCoinbaseNum uint32 = 0
			var neednextminingmsg bool = true
			for i := 0; i < int(supervene); i++ {
				msg := <-miningBlockCh
				if msg.stopKind != 0 {
					neednextminingmsg = false
				}
				if mostPowerHash == nil {
					mostPowerHashNonceBytes = msg.nonceBytes
					mostPowerHash = msg.powerHash
					mostCoinbaseNum = msg.coinbaseMsgNum
					continue
				}
				ismorepower := false
				for k := 0; k < 32; k++ {
					if msg.powerHash[k] < mostPowerHash[k] {
						ismorepower = true
						break
					} else if msg.powerHash[k] > mostPowerHash[k] {
						break
					}
				}
				if ismorepower {
					mostPowerHashNonceBytes = msg.nonceBytes
					mostPowerHash = msg.powerHash
					mostCoinbaseNum = msg.coinbaseMsgNum
				}
			}

			// 上报最大哈希结果
			uppowermsg := message.PowMasterMsg{
				Status:         message.PowMasterMsgStatusMostPowerHash,
				CoinbaseMsgNum: fields.VarInt4(mostCoinbaseNum),
				NonceBytes:     mostPowerHashNonceBytes,
				BlockHeadMeta:  inputblockheadmeta,
			}
			if neednextminingmsg {
				uppowermsg.Status = message.PowMasterMsgStatusMostPowerHashAndRequestNextMining
			}
			outputCh <- uppowermsg

			return

		}

	}(&nextstop)

}
