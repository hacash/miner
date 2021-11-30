package minerpool

import (
	"encoding/binary"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfacev2"
)

// 保存状态
func (p *MinerPool) saveStatus() error {
	stokey := []byte("status")
	value, e1 := p.status.Serialize()
	if e1 != nil {
		return e1
	}
	return p.storedb.Put(stokey, value, nil)
}

// 读取状态
func (p *MinerPool) readStatus() *MinerPoolStatus {
	statusObj := NewEmptyMinerPoolStatus()
	stokey := []byte("status")
	value, err := p.storedb.Get(stokey, nil)
	if value != nil && err == nil {
		statusObj.Parse(value, 0)
	}
	return statusObj
}

// 通过height为key保存挖出的区块hash
func (p *MinerPool) saveFoundBlockHash(height uint64, hash fields.Hash) error {
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(height))
	stokey := []byte("fdblkhx" + string(key))
	//fmt.Println("saveFoundBlockHash", height, hash.ToHex())
	err := p.storedb.Put(stokey, hash, nil)
	if err != nil {
		return err
	}
	// save status
	p.status.FindBlockHashHeightTableLastestNumber += 1
	err = p.saveStatus()
	if err != nil {
		return err
	}
	// save index
	key1 := make([]byte, 4)
	binary.BigEndian.PutUint32(key1, uint32(p.status.FindBlockHashHeightTableLastestNumber))
	val1 := make([]byte, 4)
	binary.BigEndian.PutUint32(val1, uint32(height))
	stokey1 := []byte("fdblkhx_index_" + string(key1))
	return p.storedb.Put(stokey1, val1, nil)
}

func (p *MinerPool) readFoundBlockHash(height uint64) fields.Hash {
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(height))
	stokey := []byte("fdblkhx" + string(key))
	value, err := p.storedb.Get(stokey, nil)
	if value != nil && err == nil {
		hash := fields.Hash(value)
		//fmt.Println("readFoundBlockHash", height, hash.ToHex())
		return hash
	}
	return nil // not find
}

func (p *MinerPool) readFoundBlockHashByNumber(number uint32) (uint64, fields.Hash) {
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(number))
	stokey := []byte("fdblkhx_index_" + string(key))
	value, err := p.storedb.Get(stokey, nil)
	if value != nil && err == nil && len(value) == 4 {
		height := uint64(binary.BigEndian.Uint32(value))
		hash := p.readFoundBlockHash(height)
		//fmt.Println("readFoundBlockHash", height, hash.ToHex())
		return height, hash
	}
	return 0, nil // not find
}

// 保存账户
func (p *MinerPool) saveAccountStoreData(account *Account) error {
	valuebts, e := account.storeData.Serialize()
	if e != nil {
		return e
	}
	// save
	stokey := []byte("accstodts" + string(account.address))
	err := p.storedb.Put(stokey, valuebts, nil)
	return err
}

// 读取账户
func (p *MinerPool) loadAccountStoreData(curblkhei uint64, address fields.Address) *AccountStoreData {
	stoobject := NewEmptyAccountStoreData(curblkhei)
	value, err := p.storedb.Get([]byte("accstodts"+string(address)), nil)
	if value == nil && err != nil {
		return stoobject
	}
	// parse
	stoobject.Parse(value, 0)
	return stoobject
}

// store tx
func (p *MinerPool) saveTransferTx(tx interfacev2.Transaction) error {
	p.status.TransferHashTableLastestNumber += 1
	err := p.saveStatus()
	if err != nil {
		return err
	}
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(p.status.TransferHashTableLastestNumber))
	stokey := []byte("tx_" + string(key))
	txbodys, e1 := tx.Serialize()
	if e1 != nil {
		return e1
	}
	valuebts := append([]byte{}, tx.Hash()...)
	valuebts = append(valuebts, txbodys...)
	return p.storedb.Put(stokey, valuebts, nil)
}

func (p *MinerPool) readTransferTxDataByNumber(number uint32) (fields.Hash, []byte) {
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(number))
	stokey := []byte("tx_" + string(key))
	value, err := p.storedb.Get(stokey, nil)
	if value != nil && err == nil {
		return value[0:32], value[32:]
	}
	return nil, nil
}
