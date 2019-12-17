package localcpu

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/coinbase"
	"sync"
)

// do mining
func (l *LocalCPUPowMaster) Excavate(inputblock interfaces.Block, outputblockCh chan interfaces.Block) {

	//fmt.Println(" --------------------  (l *LocalCPUPowMaster) Excavate")

	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	l.StopMining() // stop old all

	maxuint32 := uint32(4294967295)
	supervene := uint32(1)
	nonceSpace := maxuint32 / supervene

	// new run
	var nextstop byte = 0
	l.stopMarks.Store(&nextstop, &nextstop)

	//fmt.Println("	go func(stopmark *byte) {   ")

	go func(stopmark *byte) {

		var changeCoinbaseMsg uint32 = 0

	MININGLOOP:

		if changeCoinbaseMsg > 0 {
			coinbase.UpdateBlockCoinbaseMessageForMiner(inputblock, changeCoinbaseMsg)
		}
		var successMiningMark uint32 = 0
		var successFindBlock bool = false
		var syncWait = sync.WaitGroup{}
		syncWait.Add(int(supervene))

		for i := uint32(0); i < supervene; i++ {
			//fmt.Println("worker := NewCPUWorker ", i)
			worker := NewCPUWorker(&successMiningMark, outputblockCh, 0, stopmark)
			worker.coinbaseMsgNum = changeCoinbaseMsg
			//l.currentWorkers.Add( worker )
			go func(startNonce, endNonce uint32) {
				//fmt.Println( "start worker.RunMining" )
				success := worker.RunMining(inputblock, startNonce, endNonce)
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
		//fmt.Println("syncWait.Wait()  end ")
		if successFindBlock == false && *stopmark == 0 {
			//fmt.Println("changeCoinbaseMsg += 1  goto MININGLOOP ")
			changeCoinbaseMsg += 1
			goto MININGLOOP // next do mining
		}

		l.stopMarks.Delete(stopmark)

		// not find block and end it
		if successFindBlock == false {
			//fmt.Println( "start outputblockCh <- nil" )
			outputblockCh <- nil
			//fmt.Println( "end--- outputblockCh <- nil" )
		}

		// success or stop
		// return

	}(&nextstop)

}
