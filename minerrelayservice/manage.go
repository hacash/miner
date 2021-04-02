package minerrelayservice

import "github.com/hacash/miner/message"

func (r *RelayService) addClient(client *ConnClient) {
	r.changelock.Lock()
	defer r.changelock.Unlock()

	r.allconns[client.id] = client
}

func (r *RelayService) dropClient(client *ConnClient) {
	r.dropClientById(client.id)
}

func (r *RelayService) dropClientById(cid uint64) {
	r.changelock.Lock()
	defer r.changelock.Unlock()

	for id, _ := range r.allconns {
		if id == cid {
			delete(r.allconns, id)
			break
		}
	}
}

// 通知所有连接新区块到来
func (r *RelayService) notifyAllClientNewBlockStuff(blkstuff *message.MsgPendingMiningBlockStuff) {
	bts := blkstuff.Serialize()
	r.notifyAllClientNewBlockStuffByMsgBytes(bts)
}

// 通知所有连接新区块到来
func (r *RelayService) notifyAllClientNewBlockStuffByMsgBytes(stuffbts []byte) {
	r.changelock.Lock()
	defer r.changelock.Unlock()

	for _, client := range r.allconns {
		message.MsgSendToTcpConn(client.conn, message.MinerWorkMsgTypeMiningBlock, stuffbts)
	}

}
