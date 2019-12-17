package miner

func (m *Miner) doStopMining() {
	//fmt.Println("doStopMining start")
	m.stopSignCh <- true
	//fmt.Println("doStopMining end")
}
