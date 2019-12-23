package minerpool

import (
	"encoding/hex"
	"github.com/hacash/core/account"
	"github.com/hacash/core/sys"
	"log"
	"os"
	"path"
)

type MinerPoolConfig struct {
	Datadir           string
	TcpListenPort     int
	TcpConnectMaxSize uint
	FeePercentage     float64

	RewardAccount *account.Account
}

func NewEmptyMinerPoolConfig() *MinerPoolConfig {
	cnf := &MinerPoolConfig{
		TcpListenPort:     3339,
		TcpConnectMaxSize: 200,
		FeePercentage:     0.2,
	}
	return cnf
}

//////////////////////////////////////////////////

func NewMinerPoolConfig(cnffile *sys.Inicnf) *MinerPoolConfig {
	cnf := NewEmptyMinerPoolConfig()
	cnfsection := cnffile.Section("minerpool")
	defdir := path.Join(path.Dir(cnffile.MustDataDir()), ".hacash_minerpool")
	cnf.Datadir = sys.AbsDir(cnfsection.Key("data_dir").MustString(defdir))
	cnf.TcpListenPort = cnfsection.Key("listen_port").MustInt(3339)
	cnf.TcpConnectMaxSize = cnfsection.Key("max_connect").MustUint(200)
	cnf.FeePercentage = cnfsection.Key("fee_percentage").MustFloat64(0.2)
	if cnf.FeePercentage >= 1 || cnf.FeePercentage < 0 {
		panic("fee_percentage value error.")
	}
	password := cnfsection.Key("rewards_password").MustString("")
	if password == "" {
		log.Fatal("[Miner Pool Config Error] rewards password cannot be empty.")
		os.Exit(0)
	}
	var privkey []byte = nil
	if len(password) == 64 {
		key, err := hex.DecodeString(password)
		if err == nil {
			privkey = key
		}
	}
	var err error
	var reward_acc *account.Account = nil
	if privkey != nil {
		reward_acc, err = account.GetAccountByPriviteKey(privkey)
	} else {
		reward_acc = account.CreateAccountByPassword(password)
	}
	if err != nil {
		panic(err)
	}
	cnf.RewardAccount = reward_acc

	return cnf
}
