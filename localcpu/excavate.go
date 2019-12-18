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
	//maxuint32 := uint32(2294900)
	//supervene := uint32(6)
	// -------- test end   --------
	nonceSpace := maxuint32 / supervene

	// new run
	var nextstop byte = 0
	l.stopMarks.Store(&nextstop, &nextstop)

	//fmt.Println("	go func(stopmark *byte) {   ")

	go func(stopmark *byte) {

		var successMiningMark uint32 = 0
		var successFindBlock bool = false

		var successBlockCh = make(chan successBlockReturn, 1)

		var syncWait = sync.WaitGroup{}
		syncWait.Add(int(supervene))

		for i := uint32(0); i < supervene; i++ {
			//fmt.Println("worker := NewCPUWorker ", i)
			worker := NewCPUWorker(&successMiningMark, successBlockCh, 0, stopmark)
			worker.coinbaseMsgNum = l.coinbaseMsgNum
			//l.currentWorkers.Add( worker )
			go func(startNonce, endNonce uint32) {
				//fmt.Println( "start worker.RunMining" )
				success := worker.RunMining(inputblockheadmeta, startNonce, endNonce)
				if success {
					successFindBlock = true
				}
				syncWait.Done()
				//fmt.Println( "end syncWait.Done()" )
			}(nonceSpace*i, nonceSpace*i+nonceSpace)
		}

		//fmt.Println("syncWait.Wait()  start ")
		// wait
		syncWait.Wait()

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

		// success
		success := <-successBlockCh
		outputCh <- message.PowMasterMsg{
			Status:         message.PowMasterMsgStatusSuccess,
			CoinbaseMsgNum: fields.VarInt4(success.coinbaseMsgNum),
			NonceBytes:     success.nonceBytes,
			BlockHeadMeta:  inputblockheadmeta,
		}
		return // ok end

	}(&nextstop)

}
