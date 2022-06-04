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
	PoolAndWorkerAgreementVersionNumber fields.VarUint2 // Version number
	WorkerKind                          fields.VarUint1 // Mining end type
	RewardAddress                       fields.Address  // Reward collection address
}

// serialize
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

// Deserialization
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
