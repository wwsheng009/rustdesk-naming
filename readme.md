# RustDesk 自定义服务器命名规则

RustDesk Windows客户端支持通过EXE文件名嵌入自定义服务器配置信息。配置信息经过Base64编码和字符串反转处理。


## 使用方法

使用工具生成加密的配置信息，将EXE文件名命名为`rustdesk-licensed-<encoded_string>.exe`即可。

有两种工具：
- 使用sh脚本
- 使用go语言编译成工具

### 编码模式
1. 命令行参数模式:
```bash
./key.sh <key> <host> [api] [relay]
```

2. 交互模式（无参数时）:
```bash
./key.sh
```

### 解码模式
```bash
./key.sh <encoded_string>
```

## 技术实现
1. 配置信息以JSON格式存储
2. 使用Base64 URL安全编码（无填充）
3. 编码后进行字符串反转

## 示例

编码示例:
```bash
./key.sh "5Qbwsde3unUcJBtrx9ZkvUmwFNoExHzpryHuPUdqlWM=" "1.1.1.1"
```

输出:
```
rustdesk-licensed-0nI900VsFHZVBVdIlncwpHS4V0bOZ0dtVldrpVO4JHdCp0YV5WdzUGZzdnYRVjI6ISeltmIsISMuEjLx4SMiojI0N3boJye.exe
```

## 参考实现
- [RustDesk命名模块](https://github.com/rustdesk/rustdesk/blob/master/src/naming.rs)
- [自定义服务器模块](https://github.com/rustdesk/rustdesk/blob/master/src/custom_server.rs)

