# RPCg

RPCg 是一款基于 Go 标准库实现的迷你 RPC 框架，旨在**帮助我们更好的理解** RPC 框架的工作机制，模块抽象良好，易扩展和使用。 

## 事前须知

这个迷你框架是为了学习RPC框架而开发的，**无脑**重复造轮子的工作 duck 不必，但是前人为我们推导出了公式，我们为了更好地理解再去复现一遍也未尝不可，别人写的再神奇让你拍手叫绝，也终究是别人的，上手尝试才是入门的第一步。

另外，如果你有问题，欢迎随时与我交流。

**当然如果你觉得对你有帮助，给我点一个 Star 就是最大的鼓励。**

好了，话就说这么多，下面上正文。

## 特性

- 支持 TCP，UDP，Unix，HTTP 网络传输方式。
- 实现了三种序列化方式 Google Protobuf、json、gob（默认采用 Protobuf 方式）和一种压缩方式 gzip。
- 实现了六种负载均衡算法，随机，顺序选择，加权选择，基于下游服务器时延选择，基于下游服务负载的加权选择，基于一致性哈希策略。
- 定制了生成完整 RPC 代码的代码生成插件。
- 实现了自定义的注册中心，提供心跳机制维持连接。
- 实现了自定义的数据格式。
- 接口抽象良好，各模块耦合度低，模块内易扩展，网络传输、序列化器、负载均衡算法可灵活配置。

## TODO

