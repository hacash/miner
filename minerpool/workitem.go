package minerpool

import "github.com/hacash/core/interfaces"

type WorkItem struct {
	belongClient *Client
	miningBlock interfaces.Block
	miningCoinbaseMsgNum uint32
}

func NewWorkItem(cli *Client, blk interfaces.Block, num uint32) *WorkItem {
	return &WorkItem{
		belongClient:         cli,
		miningBlock:          blk,
		miningCoinbaseMsgNum: num,
	}
}



