package localgpu

// stop mining
func (l *LocalGPUPowMaster) StopMining() {
	l.stopMarks.Range(func(k interface{}, v interface{}) bool {
		mk := v.(*byte)
		*mk = 1 // set stop
		return true
	})
}
