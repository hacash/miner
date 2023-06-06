package console

import (
	"fmt"
	"github.com/hacash/miner/minerpool"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var render_cache_console string = ""

func (mc *MinerConsole) console(response http.ResponseWriter, request *http.Request) {
	if render_cache_console == "" {
		jsonstring := []string{}

		jsonstring = append(jsonstring, fmt.Sprintf(`"fee_ratio":"%.2f%%"`, mc.pool.Conf.FeePercentage*100))
		jsonstring = append(jsonstring, fmt.Sprintf(`"server_port":%d`, mc.pool.Conf.TcpListenPort))
		jsonstring = append(jsonstring, fmt.Sprintf(`"total_addresses":%d`, mc.pool.GetCurrentAddressCount()))
		jsonstring = append(jsonstring, fmt.Sprintf(`"total_clients":%d`, mc.pool.GetCurrentTcpConnectingCount()))
		jsonstring = append(jsonstring, fmt.Sprintf(`"miner_account":"%s"`, mc.pool.Conf.RewardAccount.AddressReadable))
		render_cache_console = "{" + strings.Join(jsonstring, ",") + "}"

		go func() {
			time.Sleep(time.Second * 10)
			render_cache_console = ""
		}()
	}
	mc.renderJsonString(response, render_cache_console)
}

type AccountWithPowerRatio struct {
	Account    *minerpool.Account
	PowerRatio float64
}

var data_cache_current_mining_accounts []*AccountWithPowerRatio = nil

type sortBy func(p1, p2 *minerpool.Account) bool

type accountsClientsSorter struct {
	datalist []*minerpool.Account
	by       sortBy
}

func (s *accountsClientsSorter) Len() int {
	return len(s.datalist)
}

func (s *accountsClientsSorter) Swap(i, j int) {
	s.datalist[i], s.datalist[j] = s.datalist[j], s.datalist[i]
}

func (s *accountsClientsSorter) Less(i, j int) bool {
	return s.by(s.datalist[i], s.datalist[j])
}

func (by sortBy) Sort(datalist []*minerpool.Account) {
	ps := &accountsClientsSorter{
		datalist: datalist,
		by:       by,
	}
	sort.Sort(ps)
}

func (mc *MinerConsole) addresses(response http.ResponseWriter, request *http.Request) {
	if data_cache_current_mining_accounts == nil {
		accmaps := mc.pool.GetCurrentMiningAccounts()
		acclist := []*minerpool.Account{}
		accExtList := []*AccountWithPowerRatio{}

		// sort
		if len(accmaps) > 0 {
			var totalPower uint64 = 0

			for _, v := range accmaps {
				acclist = append(acclist, v)
				totalPower += v.GetRealtimePowWorth().Uint64()
			}

			if totalPower <= 0 {
				totalPower = 10000 * 10000
			}

			by_sort_findblocks := func(a1, a2 *minerpool.Account) bool {
				b1, _ := a1.GetStoreData().GetFinds()
				b2, _ := a2.GetStoreData().GetFinds()
				return b1 > b2
			}

			sortBy(by_sort_findblocks).Sort(acclist)
			accExtList = make([]*AccountWithPowerRatio, len(acclist))

			for i, v := range acclist {
				accExtList[i] = &AccountWithPowerRatio{
					Account:    v,
					PowerRatio: float64(v.GetRealtimePowWorth().Uint64()) / float64(totalPower),
				}
			}
		}
		data_cache_current_mining_accounts = accExtList
		go func() {
			time.Sleep(time.Second * 10)
			data_cache_current_mining_accounts = nil
		}()
	}

	params := parseRequestQuery(request)

	// page limit
	var limit int = 20
	if ln, ok := params["limit"]; ok {
		if i, e := strconv.Atoi(ln); e == nil {
			limit = i
		}
	}
	if limit > 100 {
		limit = 100
	}

	var page int = 1
	if ln, ok := params["page"]; ok {
		if i, e := strconv.Atoi(ln); e == nil {
			page = i
		}
	}

	// address
	var address string = ""
	if ln, ok := params["address"]; ok {
		address = ln
		page = 1
		limit = 1
	}

	var empty_accounts = []*AccountWithPowerRatio{}

	// single row
	if address != "" {
		for _, v := range data_cache_current_mining_accounts {
			if strings.Compare(address, v.Account.GetAddress().ToReadable()) == 0 {
				renderAccountDatalist(mc, response, []*AccountWithPowerRatio{v})
				return
			}
		}
		renderAccountDatalist(mc, response, empty_accounts)
		return
	}

	// select rows
	var start = (page - 1) * limit
	var end = start + limit
	var endmax = len(data_cache_current_mining_accounts)
	if end > endmax {
		end = endmax
	}
	if start >= end {
		renderAccountDatalist(mc, response, empty_accounts)
		return
	}

	// render
	renderAccountDatalist(mc, response, data_cache_current_mining_accounts[start:end])
	return

}

func renderAccountDatalist(mc *MinerConsole, response http.ResponseWriter, accs []*AccountWithPowerRatio) {
	if len(accs) == 0 {
		mc.renderJsonString(response, `{"datalist":[]}`)
		return
	}
	jsontexts := []string{}
	for _, acc := range accs {
		jsontexts = append(jsontexts, parsePowWorkerTableRowJsonString(acc))
	}

	mc.renderJsonString(response, `{"datalist":[`+strings.Join(jsontexts, ",")+`]}`)
	return

}

func parsePowWorkerTableRowJsonString(acc *AccountWithPowerRatio) string {
	sto := acc.Account.GetStoreData()
	if sto == nil {
		return ""
	}
	f1, f2 := sto.GetFinds()
	r1, r2, r3 := sto.GetRewards()
	return fmt.Sprintf(strings.Replace(`{
		"address":"%s",
		"clients":%d,
		"realtime_power":%d,
		"realtime_power_ratio":%f,
		"find_blocks":%d,
		"find_coins":%d,
		"complete_rewards":"ㄜ%s:240",
		"deserved_rewards":"ㄜ%s:240",
		"unconfirmed_rewards":"ㄜ%s:240",
		"deserved_and_unconfirmed_rewards":"ㄜ%s:240"
	}`, "\n", "", -1),
		acc.Account.GetAddress().ToReadable(),
		acc.Account.GetClientCount(),
		acc.Account.GetRealtimePowWorth(),
		acc.PowerRatio,
		f1, f2,
		commaSplix(r1), commaSplix(r2), commaSplix(r3), commaSplix(r2+r3),
	)
}

func commaSplix(num int64) string {
	str := strconv.FormatInt(num, 10)
	sl := len(str)
	if sl > 3 {
		ns := ""
		for i := sl - 1; i >= 0; i-- {
			ns = string([]byte(str)[i]) + ns
			if (sl-i)%3 == 0 {
				ns = "," + ns
			}
		}
		return strings.TrimLeft(ns, ",")
	}
	return str
}
