package workerGPU

// to do next
func (g *GpuWorker) NextMining(nextheight uint64) {

	g.stepLock.Lock()
	defer g.stepLock.Unlock()

}
