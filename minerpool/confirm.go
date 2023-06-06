package minerpool

// Revenue recognition
func (p *MinerPool) confirmRewards(curblkheight uint64, confirmPeriod *RealtimePeriod) {
	p.periodChange.Lock()
	defer p.periodChange.Unlock()

	delayedCheckHeight := uint64(p.Conf.DelayedConfirmRewardHeight)

	if curblkheight < delayedCheckHeight*2 {
		return
	}
	for _, acc := range confirmPeriod.realtimeAccounts {
		//fmt.Println(acc.storeData.unconfirmedRewardListCount, acc.storeData.unconfirmedRewardList)
		curhei, rewards, ok := acc.storeData.unshiftUnconfirmedRewards(curblkheight - delayedCheckHeight)
		//fmt.Println("confirmRewards", ok, curhei, rewards)
		if ok {
			_ = p.saveAccountStoreData(acc)
			// check block height
			if p.checkBlockHeightMiningSuccess(uint64(curhei)) {
				//fmt.Println( "checkBlockHeightMiningSuccess:", rewards )
				if acc.storeData.moveRewards("deserved", rewards) {
				}
			}
		}
	}
}

func (p *MinerPool) checkBlockHeightMiningSuccess(height uint64) bool {

	if ok1, ok2 := p.checkBlockHeightMiningDict[height]; ok2 {
		return ok1 // cache
	}
	//
	foundblkhx := p.readFoundBlockHash(height)
	if foundblkhx == nil {
		return false
	}
	// compare
	storeHash, err := p.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead().ReadBlockHashByHeight(height)
	if err != nil {
		return false
	}
	if len(p.checkBlockHeightMiningDict) > 255 {
		p.checkBlockHeightMiningDict = map[uint64]bool{} // clean
	}
	if foundblkhx.Equal(storeHash) {
		p.checkBlockHeightMiningDict[height] = true
		return true
	} else {
		p.checkBlockHeightMiningDict[height] = false
		return false
	}
}
