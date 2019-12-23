package minerpool

func (p *MinerPool) loop() {

	for {
		select {
		case obj := <-p.successFindBlockCh:
			if obj.msg.BlockHeadMeta.GetHeight() <= p.currentSuccessFindBlockHeight {
				break
			}
			p.currentSuccessFindBlockHeight = obj.msg.BlockHeadMeta.GetHeight()
			// do add success
			obj.account.successFindNewBlock(obj.msg)
		}
	}

}
