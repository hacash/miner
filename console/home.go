package console

import (
	"fmt"
	"github.com/hacash/miner/minerpool"
	"net/http"
)

func (mc *MinerConsole) home(response http.ResponseWriter, request *http.Request) {

	// show datas at html
	htmltext := "<html><head><title>hacash miner pool home</title></head><body>"
	htmltext += `<style>#table{ border-collapse: collapse; } td{padding: 0 5px;} </style>`

	htmltext += fmt.Sprintf(`<div>
			<p>FeeRatio: %.2f %%, Addr: %s</p>
			<p>Port: %d</p>
			<p>TotalClients: %d</p>
			</div>`,
		mc.pool.Conf.FeePercentage*100,
		mc.pool.Conf.RewardAccount.AddressReadable,
		mc.pool.Conf.TcpListenPort,
		mc.pool.GetCurrentTcpConnectingCount(),
	)

	htmltext += `<table id="table" border="1">
		<tr>
			<th>#</th>
			<th>Address</th>
			<th>Clients</th>
			<th>PeriodPowWorth</th>
			<th>FindBlocks/Coins</th>
			<th>CompleteRewards</th>
			<th>DeservedRewards</th>
			<th>UnconfirmedRewards</th>
		</tr>
    	`
	curperiod := mc.pool.GetCurrentRealtimePeriod()
	if curperiod != nil {
		for i, acc := range curperiod.GetAccounts() {
			htmltext += parsePowWorkerTableRow(i+1, acc)
		}
	}
	htmltext += "</table>"

	htmltext += "</body></html>"

	// return content
	response.Write([]byte(htmltext))

}

func parsePowWorkerTableRow(num int, acc *minerpool.Account) string {
	sto := acc.GetStoreData()
	if sto == nil {
		return ""
	}
	f1, f2 := sto.GetFinds()
	r1, r2, r3 := sto.GetRewards()
	return fmt.Sprintf(`<tr>
			<td>%d</td>
			<td>%s</td>
			<td>%d</td>
			<td>%d</td>
			<td>%d/%d</td>
			<td>ㄜ%s:240</td>
			<td>ㄜ%s:240</td>
			<td>ㄜ%s:240</td>
		</tr>`,
		num,
		acc.GetAddress().ToReadable(),
		acc.GetClientCount(),
		acc.GetRealtimePowWorth(),
		f1, f2,
		commaSplix(r1), commaSplix(r2), commaSplix(r3),
	)
}
