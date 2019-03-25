# TimeForCoin.Server

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[![Golang Version](https://img.shields.io/badge/Golang-1.12.1-blue.svg)](https://golang.org/doc/devel/release.html#go1.12)
[![GoDoc](https://godoc.org/github.com/TimeForCoin/Server/app?status.svg)](https://godoc.org/github.com/TimeForCoin/Server/app)

[![](https://api.travis-ci.org/TimeForCoin/Server.svg?branch=master)](https://www.travis-ci.org/TimeForCoin/Server)
[![codecov](https://codecov.io/gh/TimeForCoin/Server/branch/master/graph/badge.svg)](https://codecov.io/gh/TimeForCoin/Server)

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/52975f5b0d5e4b7aa257601e79919384)](https://app.codacy.com/app/ZhenlyChen/Server?utm_source=github.com&utm_medium=referral&utm_content=TimeForCoin/Server&utm_campaign=Badge_Grade_Dashboard)
[![CodeFactor](https://www.codefactor.io/repository/github/timeforcoin/server/badge)](https://www.codefactor.io/repository/github/timeforcoin/server)
[![codebeat badge](https://codebeat.co/badges/9b89186f-55a9-42a4-bc6f-f6f10c36c1cd)](https://codebeat.co/projects/github-com-timeforcoin-server-master)

TimeForCoin 服务端

## Usage 用法

须使用`Golang 1.12.x`以上的版本(或者使用`1.11.x`并启用`mod`功能)

```bash
$ git clone https://github.com/TimeForCoin/Server.git
$ go mod download
```

按照`config.default.yaml`创建配置文件，运行应用(默认从当前目录读取配置文件`config.yaml`)

```bash
$ go run main.go -c ./config.yaml
```



## Test 测试

提交前建议先运行并通过单元测试

设置测试需要的环境变量，然后执行测试

```bash
$ go test -v .\app\...
```

`Windows`下的测试可以直接使用`powershell`执行 (须先设置好配置)

```bash
$ ./test.default.ps1
```



## Deploy 部署

默认使用`Docker`进行部署

```bash
$ docker build -t time-for-coin:v1 .
$ docker run -p 30233:30233 --name my-time-for-coin -d time-for-coin:v1
```

