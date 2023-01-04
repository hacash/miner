# Golang Hacash Full Node Miner

### Software download & release log

Release: [software_release_log.md](doc/software_release_log.md)

### Compilation build instructions 

Compilation build instructions: [build_compilation_en.md](doc/build_compilation_en.md)

---

### RPC API Doc 

English：[rpc_api_doc.en.md](https://github.com/hacash/service/blob/master/doc/rpc_api_doc.en.md) 

Chinese：[rpc_api_doc.cn.md](https://github.com/hacash/service/blob/master/doc/rpc_api_doc.cn.md) 

---

### Miner service Doc

English： [miner_service_api.en.md](https://github.com/hacash/service/blob/master/doc/miner_service_api.en.md)

Chinese：[miner_service_api.cn.md](https://github.com/hacash/service/blob/master/doc/miner_service_api.cn.md) 


---

### X16RS algorithm design explanation

English： [x16rs_algorithm_description.en.md](https://github.com/hacash/x16rs/blob/master/doc/x16rs_algorithm_description.en.md)

Chinese：[x16rs_algorithm_description.cn.md](https://github.com/hacash/x16rs/blob/master/doc/x16rs_algorithm_description.cn.md)


---

### HAC & HACD mining fairness description

English： [HAC_HACD_mining_fairness_description.en.md](https://github.com/hacash/x16rs/blob/master/doc/HAC_HACD_mining_fairness_description.en.md)

Chinese：[HAC_HACD_mining_fairness_description.cn.md](https://github.com/hacash/x16rs/blob/master/doc/HAC_HACD_mining_fairness_description.cn.md)



---

### Project code engineering architecture

The architecture of Hacash full node code is divided into 7 levels from bottom to top:

X16RS -> Core -> Chain -> Mint -> Node -> Service -> Miner

Each layer of the architecture has independent functions and responsibilities for the upper layer to call, and the implementation of the lower layer to the upper layer is unknown. The responsibilities of each layer are roughly as follows:

1. [X16RS] Basic algorithm - including HAC mining, block diamond mining, GPU version algorithm, etc.
2. [Core] Core - block structure definition, interface definition, data serialization and deserialization, storage object, field format, genesis block definition, etc.
3. [Chain] Chain - underlying database, block and transaction storage, blockchain state storage, logs, etc.
4. [Mint] Mint - block mining difficulty adjustment algorithm, coinbase definition, block construction, transaction execution and status update, etc.
5. [Node] Node - P2P underlying module, Backend blockchain synchronization terminal, point-to-point network message definition and processing, etc.
6. [Service] Service - RPC API interface service, block and transaction and account data query, other services, etc.
7. [Miner] Miner - block construction and mining, diamond mining, transaction memory pool, mining pool server, mining pool worker, etc.


```cgo
// compile NOTE: set go module env config
GO111MODULE="auto"
```
