package minerpool

import (
	"encoding/binary"
	"github.com/hacash/core/fields"
)

// 保存状态
func (p *MinerPool) saveStatus() error {
	return nil
}

// 读取状态
func (p *MinerPool) readStatus() error {
	return nil
}

// 通过height为key保存挖出的区块hash
func (p *MinerPool) saveFoundBlockHash(height uint64, hash fields.Hash) error {
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(height))
	stokey := []byte("fdblkhx" + string(key))
	return p.storedb.Put(stokey, hash, nil)
}

func (p *MinerPool) readFoundBlockHash(height uint64) fields.Hash {
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(height))
	stokey := []byte("fdblkhx" + string(key))
	value, err := p.storedb.Get(stokey, nil)
	if value != nil && err == nil {
		return fields.Hash(value)
	}
	return nil // not find
}

// 保存账户
func (p *MinerPool) saveAccountStoreData(account *Account) error {
	valuebts, e := account.storeData.Serialize()
	if e != nil {
		return e
	}
	// save
	stokey := []byte("accstodts" + string(account.address))
	return p.storedb.Put(stokey, valuebts, nil)
}

// 读取账户
func (p *MinerPool) loadAccountStoreData(address fields.Address) *AccountStoreData {
	stoobject := NewEmptyAccountStoreData()
	value, err := p.storedb.Get([]byte("accstodts"+string(address)), nil)
	if value == nil && err != nil {
		return stoobject
	}
	// parse
	stoobject.Parse(value, 0)
	return stoobject
}
