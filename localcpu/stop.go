package localcpu

// stop mining
func (l *LocalCPUPowMaster) StopMining() {
	l.stopMarks.Range(func(k interface{}, v interface{}) bool {
		mk := v.(*byte)
		*mk = 1 // set stop
		return true
	})
}
