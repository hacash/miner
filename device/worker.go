package device

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	itfcs "github.com/hacash/miner/interfaces"
	"sync"
)

type PoWWokerMng struct {
	config           itfcs.PoWConfig
	device           itfcs.PoWDevice
	execLock         sync.Mutex
	stopMarks        sync.Map
	isUploadHashrate bool
}

func NewPoWWorkerMng(alloter itfcs.PoWExecute) *PoWWokerMng {
	var dvs = NewPoWDeviceMng(alloter)
	return &PoWWokerMng{
		config:           alloter.Config(),
		device:           dvs,
		execLock:         sync.Mutex{},
		stopMarks:        sync.Map{},
		isUploadHashrate: true,
	}
}

func (c *PoWWokerMng) Config() itfcs.PoWConfig {
	return c.config
}

func (w *PoWWokerMng) CloseUploadHashrate() {
	w.isUploadHashrate = false
}

func (w *PoWWokerMng) Init() error {
	return w.device.Init()
}

func (w *PoWWokerMng) StopMining() {
	w.stopMarks.Range(func(k interface{}, v interface{}) bool {
		mk := v.(*byte)
		*mk = 1 // set stop
		return true
	})
	w.device.StopMining()
}

func (w *PoWWokerMng) createNewBrief(stuff *itfcs.PoWStuffOverallData,
	blk_nonce fields.VarUint4, cb_nonce []byte) *itfcs.PoWStuffBriefData {
	if cb_nonce == nil {
		cb_nonce = make([]byte, 32)
		rand.Read(cb_nonce)
		//fmt.Println(" *** createNewBrief cb_nonce = ", hex.EncodeToString(cb_nonce))
	}
	// replace
	cbtx := stuff.CoinbaseTx.CopyForMining()
	//fmt.Println("COinbase ", cbtx.Address.ToReadable(), cbtx.ExtendDataVersion)
	cbtx.MinerNonce = cb_nonce
	var cb_hash = cbtx.Hash()
	// mklrRoot
	var useblk = stuff.BlockHeadMeta.CopyHeadMetaForMining()
	mklr_root := blocks.CalculateMrklRootByCoinbaseTxModify(cb_hash, stuff.MrklCheckTreeList)
	useblk.SetMrklRoot(mklr_root)
	//fmt.Println("mklr_root ", hex.EncodeToString(cb_nonce), hex.EncodeToString(cb_hash), mklr_root.ToHex())
	//fmt.Println("stuff.CoinbaseTx.ExtendDataVersion-----------", stuff.CoinbaseTx.ExtendDataVersion, hex.EncodeToString(cb_nonce), cb_hash.ToHex(), mklr_root.ToHex())
	// nonce
	if blk_nonce > 0 {
		useblk.SetNonce(uint32(blk_nonce))
	}
	// ok
	return &itfcs.PoWStuffBriefData{
		BlockHeadMeta: useblk,
		CoinbaseNonce: fields.Hash(cb_nonce),
	}
}

// find block
func (w *PoWWokerMng) DoMining(input *itfcs.PoWStuffOverallData) (*itfcs.PoWResultData, error) {

	w.execLock.Lock()
	defer w.execLock.Unlock()

	var stopmark byte = 0
	w.stopMarks.Store(&stopmark, &stopmark)

	var stopMarkCh = make(chan bool)
	var briefStuffCh = make(chan *itfcs.PoWStuffBriefData)

	// create mining stuff
	go func() {
		for {
			select {
			case <-stopMarkCh:
				close(stopMarkCh)
				return // close
			case briefStuffCh <- w.createNewBrief(input, 0, nil):
				// next
			}
		}
	}()

	var result *itfcs.PoWResultData = nil
	var err error = nil

	// do mining
STARTMINING:
	res, e1 := w.device.DoMining(&stopmark, briefStuffCh)
	if e1 != nil {
		err = e1
	} else if res == nil {
		err = fmt.Errorf("w.device.DoMining return nil")
	} else {
		if result == nil || bytes.Compare(result.ResultHash, res.ResultHash) == 1 {
			result = res
		}
		// check is SUCCESS find block !!!
		if stopmark == 1 {
			// end this block height upload hashrate
		} else if res.FindSuccess.Check() {
			// check nonce
			checkobj := w.createNewBrief(input, res.BlockNonce, res.CoinbaseNonce)
			blkhx := checkobj.BlockHeadMeta.HashFresh()
			if blkhx.Equal(res.ResultHash) {
				// SUCCESS
				stopmark = 1
				w.StopMining() // stop all
				// SUCCESS END
			} else {
				fmt.Printf("\n--------\nDevice find block fail, need hx %s but got %s \n--------\n",
					res.ResultHash.ToHex(), blkhx.ToHex())
				// next do mining
				goto STARTMINING
			}
		} else {
			// next do mining
			goto STARTMINING
		}
	}

	// close
	stopMarkCh <- true
	close(briefStuffCh)
	w.stopMarks.Delete(&stopmark)

	// upload
	if w.isUploadHashrate {
		return result, err
	}

	// not upload
	return nil, err
}
