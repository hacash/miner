/***
+-----------------------------------------------------------------------------------------------------------------------------------------------+
| RE              : relayserviceconfig_test.go
|                   Hacash Relay Service configuration Test code
| DATE            : Aug-10-2021
| AUTHOR          : MKD
| LOCATION        : hacash/miner/minerrelayservice
| BUILD STEPS     : go mod init relayserviceconfig_test
|                   go mod tidy
| TESTING         : go test
|                   go test -v -race
|                   go test -cover
| NOTES           : From Go V1.11 you need to use 'go mod ...', 'go mod tidy' before running 'go test'
|                   I have provided various forms of actual 'go test ...', any one of which should be sufficient to test
|                   Relay Services configuration
|                   Post testing NOTE: Once testing is complete, remove the files "go.mod" and "go.sum" to clean up files used in testing
| KNOWN PROBLEMS  :	Unable to figure out how to handle cnf.ServerAddress, aka TCPAddr
|
| CHANGE LOG      : Version       Date                                              Description of code change
|                   -------    -----------    --------------------------------------------------------------------------------------------------
|                   0.1.1      Aug-12-2021    Renamed source file, from [relayservice_test.go] to [relayserviceconfig_test.go]
|                   0.1.0      Aug-10-2021    Initial inception
+-----------------------------------------------------------------------------------------------------------------------------------------------+
*/


package minerrelayservice


import (
	"fmt"
	"github.com/hacash/core/sys"
	"os"
	"log"
	"path/filepath"
	"testing"
)


// 验证 [/tmp] 存在
// Verify [/tmp] exists
func isTmpDir() error {
    const tmpDir string = "/tmp"

    _, err := os.Stat(tmpDir)

    return err
}


// 在 /tmp 中创建服务中继配置文件
// Create the Service Relay config file in /tmp
func createConfigFile(tmpDir string, configFile string) (*os.File, error) {
    dst, err := os.Create(filepath.Join(tmpDir, filepath.Base(configFile)))

    return dst, err
}


// 使用 [Service Relay] 相关对象填充 Service Relay 配置文件
// Populate the Service Relay config file with [Service Relay] related objects
func populateConfigFile(dst *os.File, relayServiceConfig []string) {
    for _, line := range relayServiceConfig {
        fmt.Fprintln(dst, line)
    }
}


// Process the settings of the service Relay profile for further processing
// Handle the setup of the Service Relay config file for further processing
func handleConfigFile(relayServiceConfig []string) {
    const tmpDir string = "/tmp"
    const configFile string = "/relayservice.config.ini"

    itd_err := isTmpDir()

    if os.IsNotExist(itd_err) {
        log.Fatal("isTmpDir: Directory /tmp does not exist.")
    }

    dst, ccf_err := createConfigFile(tmpDir, configFile)
    defer dst.Close()

    if ccf_err != nil {
        log.Fatal("createConfigFile: Failed to create relay service file in path /tmp")
    }

    populateConfigFile(dst, relayServiceConfig)
}


// Function that returns the path of the absolute service relay configuration file
// Function that returns the absolute Service Relay config file path
func returnAbsConfigFilePath(tmpDir string, configFile string) string {
    const separator string = "/"

    return tmpDir + separator + configFile
}


// Functions to run the test suite
// Function that runs the test suite
func TestNewMinerRelayServiceConfig_001(t *testing.T) {
    const tmpDir          string = "/tmp"
    const configFile      string = "relayservice.config.ini"

    // Actual relay service content
    // Actual Relay Service contents
    relayServiceConfig := []string {"server_connect = 127.0.0.1:3350",
                                    "server_listen_port = 19991",
                                    "http_api_listen_port = 8080",
                                    "accept_hashrate = true",
                                    "report_hashrate = true",
                                    "[store]",
                                    "enable = true",
                                    "data_dir = ./hacash_relay_service_data",
                                    "save_mining_block_stuff = true",
                                    "save_mining_hash = true",
                                    "save_mining_nonce = true"}

        // Single parameter to test
        // Individual parameters to be tested
	const serverConnectPort         int =  3350
	const serverListenPort          int = 19991
	const httpApiListenPort         int =  8080
	const accepthHashrate          bool =  true
	const reportHashrate           bool =  true
	const storeEnable              bool =  true
	const storeDataDir           string =  "./hacash_relay_service_data"
	const storeSMH                 bool =  true
	const storeSMN                 bool =  true


    // The following code block does the following:
    // 1-在路径/tmp中创建Service Relay配置文件
    // 2-initialize the service relay environment to allow testing
    // 3- run the test to verify whether the service trunk code is running normally

    // The following block of code performs the following:
    // 1- Create the Service Relay configuration file in the path /tmp
    // 2- Initialize the Service Relay environment to allow testing
    // 3- Run the tests to verify the Service Relay code is functioning correctly


    // The following code block creates and populates the service relay configuration file
    // The following block of code creates then populates the Service Relay config file
    handleConfigFile(relayServiceConfig)

    // The following code block provides us with the fully qualified absolute location of the service relay configuration file
    // The following block of code gives us the fully qualified, absolute location of the Service Relay config file
    abs_target_ini_file := returnAbsConfigFilePath(tmpDir, configFile)

    target_ini_file     := sys.AbsDir(abs_target_ini_file)

    hinicnf, _          := sys.LoadInicnf(target_ini_file)

    // The following code block provides us with the fully qualified absolute location of the service relay configuration file
    // The following block of code initializes the Service Relay environment
    cnf                 := NewMinerRelayServiceConfig(hinicnf)


    // The following code block performs a test to confirm that the service relay setting parameters are correct
    // The following block of code performs the tests to confirm Service Relay set parameters correctly
    if (cnf.ServerAddress).Port != serverConnectPort {
        t.Error("Test Failed: Defined config value [ServerAddress.Port] {}, returned value {}", serverConnectPort, (cnf.ServerAddress).Port)
    }

	if cnf.ServerTcpListenPort != serverListenPort {
		t.Error("Test Failed: Defined config value [listenPort] {}, returned value {}", serverListenPort, cnf.ServerTcpListenPort)
	}

	if cnf.HttpApiListenPort != httpApiListenPort {
		t.Error("Test Failed: Defined config value [HttpApiListenPort] {}, returned value {}", httpApiListenPort, cnf.HttpApiListenPort)
	}

	if cnf.IsAcceptHashrate != accepthHashrate {
		t.Error("Test Failed: Defined config value [accepthHashrate] {}, returned value {}", accepthHashrate, cnf.IsAcceptHashrate)
	}

	if cnf.IsReportHashrate != reportHashrate {
		t.Error("Test Failed: Defined config value [reportHashrate] {}, returned value {}", reportHashrate, cnf.IsReportHashrate)
	}

	if cnf.StoreEnable != storeEnable {
		t.Error("Test Failed: Defined config value [storeEnable] {}, returned value {}", storeEnable, cnf.StoreEnable)
	}

	if cnf.DataDir != storeDataDir {
		t.Error("Test Failed: Defined config value [storeDataDir] {}, returned value {}", storeDataDir, cnf.DataDir)
	}

	if cnf.SaveMiningHash != storeSMH {
		t.Error("Test Failed: Defined config value [storeSMH] {}, returned value {}", storeSMH, cnf.SaveMiningHash)
	}

	if cnf.SaveMiningNonce != storeSMN {
		t.Error("Test Failed: Defined config value [storeSMN] {}, returned value {}", storeSMN, cnf.SaveMiningNonce)
	}
}

// End of code relayserviceconfig_test.go
