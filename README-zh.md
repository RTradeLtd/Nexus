# 🦑 Nexus

> IPFS网络节点编排协调工具

Nexus是 [IPFS](https://github.com/ipfs/go-ipfs)私有网络的节点编排，与[Temporal](https://github.com/RTradeLtd/Temporal)服务注册的工具。Nexus可以为您提供按需部署、资源管理、元数据持久化，以容器Docker方式运行IPFS节点等细粒度功能。

[![GoDoc](https://godoc.org/github.com/RTradeLtd/Nexus?status.svg)](https://godoc.org/github.com/RTradeLtd/Nexus)
[![Build Status](https://travis-ci.com/RTradeLtd/Nexus.svg?branch=master)](https://travis-ci.com/RTradeLtd/Nexus)
[![codecov](https://codecov.io/gh/RTradeLtd/Nexus/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/Nexus)
[![Go Report Card](https://goreportcard.com/badge/github.com/RTradeLtd/Nexus)](https://goreportcard.com/report/github.com/RTradeLtd/Nexus)
[![Latest Release](https://img.shields.io/github/release/RTradeLtd/Nexus.svg?colorB=red)](https://github.com/RTradeLtd/Nexus/releases)

## 多语言

[![](https://img.shields.io/badge/Lang-English-blue.svg)](README.md)  [![jaywcjlove/sb](https://jaywcjlove.github.io/sb/lang/chinese.svg)](README-zh.md)

## 安装和使用

```bash
$> go get -u github.com/RTradeLtd/Nexus/cmd/nexus
```

发布版本下载：[Releases](https://github.com/RTradeLtd/Nexus/releases) 。使用默认配置，运行Nexus的守护进程：

```bash
$> nexus init
$> nexus daemon
```

可以通过`nexus -help`命令获得更多使用说明帮助，配置文档生成`init`命令可以在 [configuration 源码](https://github.com/RTradeLtd/Nexus/blob/master/config/config.go)中找到。

## 开发

项目环境要求 [Docker CE](https://docs.docker.com/install/#supported-platforms)
以及不低于[Go 1.11](https://golang.org/dl/) 版本。

可以克隆此仓库或者使用`go get`来获取基础代码:

```bash
$> go get github.com/RTradeLtd/Nexus
```

开发者也可以直接使用make命令进行构建

```bash
$> make   # installs dependencies and builds a binary
```

### 测试

执行测试脚本的时候，请确保docker环境已经运行

```bash
$> make test
```

之后，可以使用`make clean`命令删除剩余资产。

### 本地运行

使用下面一些make命令可以轻松模拟机器上的编排环境：

```bash
$> make dev-config # make sure dev configuration is up to date
$> make testenv    # initialize test environment
$> make daemon     # start up daemon with dev configuration
```

然后，你可以启动一个网络节点：

```bash
$> make new-network   # create network entry in database
$> make start-network # spin up network node
```

### ctl

目前仍处于实验性质的轻量gRPC API控制器可通过以下执行`nexus ctl`方式获得，依赖以公开客户端 [ctl](https://github.com/bobheadxi/ctl)的形式暴露。

```bash
$> nexus ctl help
$> nexus -dev ctl StartNetwork Network=test-network
$> nexus -dev ctl NetworkStats Network=test-network
$> nexus -dev ctl StopNetwork Network=test-network
```
