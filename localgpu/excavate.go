package localgpu

import (
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/message"
	"sync"
)

// do mining
func (l *LocalGPUPowMaster) Excavate(inputblockheadmeta interfaces.Block, outputCh chan message.PowMasterMsg) {

	//fmt.Println(" --------------------  (l *LocalGPUPowMaster) Excavate")

	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	l.StopMining() // stop old all

	supervene := uint32(l.config.Concurrent) // 并发线程
	// -------- test start --------
	//maxuint32 = uint32(2294900)
	//supervene := uint32(6)
	// -------- test end   --------

	// new run
	var nextstop byte = 0
	l.stopMarks.Store(&nextstop, &nextstop)

	//fmt.Println("	go func(stopmark *byte) {   ")

	go func(stopmark *byte) {

		defer func() {
			//fmt.Println("	l.stopMarks.Delete(&nextstop) // clean  ")
			l.stepLock.Lock()
			l.stopMarks.Delete(&nextstop) // clean
			l.stepLock.Unlock()
		}()

		var successMiningMark uint32 = 0
		var successFindBlock bool = false

		var miningBlockCh = make(chan miningBlockReturn, 1)
		if l.config.ReturnPowerHash {
			miningBlockCh = make(chan miningBlockReturn, supervene)
		}
		var syncWait = sync.WaitGroup{}
		syncWait.Add(int(supervene))

		//fmt.Println("worker := NewCPUWorker ", i)
		worker := NewGPUWorker(&successMiningMark, miningBlockCh, 0, stopmark, l.config)
		if l.config.ReturnPowerHash {
			worker.returnPowerHash = true
		}
		l.StopMining()
		//fmt.Println("g.StopAllMining()")

		// Serial synchronization
		l.stepLock.Lock()
		defer l.stepLock.Unlock()

		// stop mark
		var stopmark1 byte = 0
		l.stopMarks.Store(&stopmark1, &stopmark1)
		defer l.stopMarks.Delete(&stopmark1) //l.currentWorkers.Add( worker )

		//fmt.Println( "start worker.RunMining" )
		success := worker.RunMining(inputblockheadmeta, &stopmark1)
		//fmt.Println( "end worker.RunMining", success )
		if success {
			successFindBlock = true
		}
		syncWait.Done()

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
				CoinbaseMsgNum: fields.VarUint4(success.coinbaseMsgNum),
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
					CoinbaseMsgNum: fields.VarUint4(l.coinbaseMsgNum),
					BlockHeadMeta:  inputblockheadmeta,
				}
				return
			}

		} else {

			// Statistics and comparison of mining hash results
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

			// Report the maximum hash result
			uppowermsg := message.PowMasterMsg{
				Status:         message.PowMasterMsgStatusMostPowerHash,
				CoinbaseMsgNum: fields.VarUint4(mostCoinbaseNum),
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
