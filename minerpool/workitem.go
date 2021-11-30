package minerpool

import "github.com/hacash/core/interfacev2"

type WorkItem struct {
	belongClient         *Client
	miningBlock          interfacev2.Block
	miningCoinbaseMsgNum uint32
}

func NewWorkItem(cli *Client, blk interfacev2.Block, num uint32) *WorkItem {
	return &WorkItem{
		belongClient:         cli,
		miningBlock:          blk,
		miningCoinbaseMsgNum: num,
	}
}
