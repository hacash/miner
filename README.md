# Golang Hacash Full Node Miner

### Software download & release log (软件下载和版本发布日志)

Release: [software_release_log.md](doc/software_release_log.md)

### Compilation build instructions (编译部署文档)

Compilation build instructions: [build_compilation_en.md](doc/build_compilation_en.md)

---

### RPC API Doc (区块链数据接口文档)

中文：[rpc_api_doc.cn.md](https://github.com/hacash/service/doc/rpc_api_doc.cn.md) 

English：[rpc_api_doc.en.md](https://github.com/hacash/service/doc/rpc_api_doc.en.md) 

---

### Miner service Doc (挖矿与矿池开发文档)

中文：[miner_service_api.cn.md](https://github.com/hacash/service/doc/miner_service_api.cn.md) 

English： `waiting for translation`

---

### Project code engineering architecture (项目代码工程基础架构)

Hacash 全节点代码的架构从底至上分为 7 个层级：

X16RS -> Core -> Chain -> Mint -> Node -> Service -> Miner

架构的每一层各自具备独立的功能和职责以供上层调用，而下层对上层的实现未知。其各层的职责大略如下：

1. [X16RS] 基础算法 - 包含HAC挖掘、区块钻石挖掘、GPU版本算法等
2. [Core] 核心 - 区块结构定义、Interface定义、数据序列化及反序列化、储存对象、各字段格式、创世区块定义等
3. [Chain] 链 - 底层数据库、区块和交易储存器、区块链状态储存、日志等
4. [Mint] 造币厂 - 区块挖掘难度调整算法、coinbase定义、区块构建及交易执行和状态更新等
5. [Node] 节点 - P2P底层模块、Backend区块链同步端、点对点网络消息定义及处理等
6. [Service] 服务 - RPC API 接口服务、区块和交易和账户数据等查询、其它服务等
7. [Miner] 矿工 - 区块构建及挖掘、钻石挖掘、交易内存池、矿池服务端、矿池worker等

