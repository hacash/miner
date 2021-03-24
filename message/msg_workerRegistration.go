package message

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**

挖矿工人登记

*/

const MsgWorkerRegistrationSize = 2 + 1 + 21

type MsgWorkerRegistration struct {
	PoolAndWorkerAgreementVersionNumber fields.VarUint2 // 版本号
	WorkerKind                          fields.VarUint1 // 挖矿端类型
	RewardAddress                       fields.Address  // 收取奖励地址
}

// 序列化
func (m MsgWorkerRegistration) Serialize() []byte {

	buf := bytes.NewBuffer([]byte{})
	b1, _ := m.PoolAndWorkerAgreementVersionNumber.Serialize()
	b2, _ := m.WorkerKind.Serialize()
	b3, _ := m.RewardAddress.Serialize()
	buf.Write(b1)
	buf.Write(b2)
	buf.Write(b3)

	return buf.Bytes()
}

// 反序列化
func (m *MsgWorkerRegistration) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error = nil
	seek, e = m.PoolAndWorkerAgreementVersionNumber.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.WorkerKind.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.RewardAddress.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}
