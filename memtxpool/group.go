package memtxpool

import (
	"github.com/hacash/core/fields"
	"sync"
)

type TxGroup struct {
	Head        *TxItem
	Tail        *TxItem
	itemsLocker sync.Mutex
	items       map[string]*TxItem
	Count       int64
}

func NewTxGroup() *TxGroup {
	return &TxGroup{
		items: make(map[string]*TxItem),
	}
}

//////////////////////////////////////////////////////////////////////

func (g *TxGroup) GetItem(id string) (*TxItem, bool) {
	g.itemsLocker.Lock()
	defer g.itemsLocker.Unlock()

	i, h := g.items[id]
	return i, h
}

func (g *TxGroup) Clean() {
	g.itemsLocker.Lock()
	defer g.itemsLocker.Unlock()

	g.Head = nil
	g.Tail = nil
	g.items = make(map[string]*TxItem)
	g.Count = 0
}

func (g *TxGroup) Add(item *TxItem) bool {

	g.itemsLocker.Lock()
	defer g.itemsLocker.Unlock()

	key := string(item.hash)
	if _, ok := g.items[key]; ok == false {
		if g.Count == 0 {
			g.Head = item
			g.Tail = item
		} else {
			previtem := g.Tail.prev
			curitem := g.Tail
			for {
				if item.feepurity <= g.Tail.feepurity { // is tail
					item.prev = g.Tail
					g.Tail.next = item
					g.Tail = item
					break
				} else if item.feepurity == curitem.feepurity { // insert after
					if item.tx.GetFee().MoreThan(curitem.tx.GetFee()) {
						// 手续费值大于，则排在前面// insert before
						previtem.next = item
						curitem.prev = item
						item.next = curitem
						item.prev = previtem
						break
					} else {
						// 手续费含量相同，但费用实际值小于或等于，则排在后面
						oldnext := curitem.next
						curitem.next = item
						item.prev = curitem
						oldnext.prev = item
						item.next = oldnext
						break
					}
				} else if item.feepurity > curitem.feepurity {
					if previtem == nil { // is head
						item.next = g.Head
						g.Head.prev = item
						g.Head = item
						break
					} else if item.feepurity < previtem.feepurity { // insert before
						previtem.next = item
						curitem.prev = item
						item.next = curitem
						item.prev = previtem
						break
					}
				}
				// check prev
				curitem = curitem.prev
				if curitem == nil { // is head
					item.next = g.Head
					g.Head.prev = item
					g.Head = item
					break
				}
				previtem = curitem.prev
			}
		}
		g.items[key] = item
		g.Count += 1
		return true
	}
	return false
}

func (g *TxGroup) Find(hash fields.Hash) *TxItem {
	g.itemsLocker.Lock()
	defer g.itemsLocker.Unlock()

	if havtx, ok := g.items[string(hash)]; ok {
		return havtx
	}
	return nil
}

func (g *TxGroup) RemoveByTxHash(hash fields.Hash) *TxItem {

	g.itemsLocker.Lock()
	defer g.itemsLocker.Unlock()

	key := string(hash)
	if havitem, ok := g.items[key]; ok {
		g.RemoveItem(havitem)
		return havitem
	}
	return nil
}

func (g *TxGroup) RemoveItem(item *TxItem) bool {

	g.itemsLocker.Lock()
	defer g.itemsLocker.Unlock()

	key := string(item.hash)
	if havtx, ok := g.items[key]; ok {
		if g.Count == 1 {
			g.Head = nil
			g.Tail = nil
		} else if havtx == g.Head {
			g.Head = g.Head.next
			g.Head.prev = nil
		} else if havtx == g.Tail {
			g.Tail = g.Tail.prev
			g.Tail.next = nil
		} else {
			havtx.prev.next = havtx.next
			havtx.next.prev = havtx.prev // drop
		}
		delete(g.items, key)
		g.Count -= 1
		return true
	}
	return false
}
