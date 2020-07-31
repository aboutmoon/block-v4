# block-v4

#####1. 将创建区块链的操作放到命令
- NewBlockChain
#####2. NewBlockChain函数的重写
- a. 交易ID
- b. 交易输入: TXInput
- c. 交易输出: TXOutput
#####3. 根据交易结构，改写代码
- a. 创建区块链的时候生成奖励
- b. 通过指定地址检索到他相关的UTXO
- c. 实现UTXO的转移（创建交易函数: NewTransaction（from,to,amount））

 