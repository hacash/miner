package interfaces

import "github.com/hacash/core/fields"

type HashratePool interface {
	NewMiningStuff(*PoWStuffOverallData)
	ReportHashrate(fields.Address, *PoWStuffOverallData, *PoWResultShortData)
}
