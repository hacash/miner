package miningpool

import "github.com/hacash/core/fields"

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
	return nil
}

// 保存账户
func (p *MinerPool) saveAccount(account *Account) error {
	return nil
}

// 读取账户
func (p *MinerPool) readAccount(address fields.Address) *Account {

	return nil
}
