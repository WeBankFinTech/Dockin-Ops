# Dockin Ops - Dockin Operation service

[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

[English](README.md) | 中文

Dockin运维管理系统是安全的运维管理服务，优化exec执行性能、支持命令权限管理、支持场景运维编排。

**更多Dockin组件请访问 [https://github.com/WeBankFinTech/Dockin](https://github.com/WeBankFinTech/Dockin)**

![Architecture](docs/images/dockin.png)

## 组件介绍
### Dockin-Opserver
dockin-opserver是基于kubernetes client-go开发的apiserver接口代理，支持如下基本功能：
- 多集群管理，可以同时管理多套apiserver集群
- 用户账号信息管理，Pod本身访问时没有账号管理的，因此只要具有kubeconfig就可以访问所有Pod，对跨部门等情况来说，存在安全隐患
- ssh代理，用户通过kubectl exec /bin/bash -it的方式登录到Pod中之后，可以操作任何命令，但是由于kubernetes的特性，对memory存在严格的控制，因此一些耗费memory的操作，比如vi一个特大文件，就很容易造成Pod被OOM kill，通过ssh代理，我们可以拦截所有用户执行的命令，并且进行安全判断（黑白名单）。
- 审计，通过kubectl exec的方式可以执行任何命令，那如何用户执行了何种命令而导致了安全隐患，通过审计功能，我们将用户执行的所有命令，无论是exec还是在shell环境中，都有相对应的存档，做到有迹可循。
- 提供http和websocket接口，用于执行普通exec和交互式exec的请求
- 协议转换，将websocket和spdy协议数据互相转换
- 通过client-go的informer功能，将pod的Add、Update、Delete的时间保存到redis中

### Dockin-Opsctl
类似kubectl客户端，二进制客户端，用户和dockin-opserver建立http或者websocket请求，绑定当前标准输入和标准输入，在交互式模式下，进入raw模式

### Dockin-Opagent
dockin的agent，通过daemonset的方式部署在各个kubernetes节点中，主要有如下功能
- 挂载docker.sock，用于连接dockerd
- 集成docker api，进行docker exec操作
- 管理当前节点中containerId和podName的对应关系
- 提供spdy接口，用于响应dockin-opserver发起的exec请求
- 绑定docker api的输入输出流到spdy的输入输出流

### dockctl
dockctl为dockin-opsctl的一个包装
- dockctl提供了批量操作多个Pod，以子系统为单位
- 美化输出, dockin-opsctl返回的都是标准输出，大多为json数据, dockctl会将json数据进行美化展示

## 已开源功能列表
- dockctl cmdb，基于子系统限制Pod相关信息
- exec代理
- ssh工具
  - 支持拦截ssh命令
  - 支持命令参数拦截
  - 支持账号管理
- Pod权限管理

## Roadmap
- shell内容解析优化（基于逃逸字符、控制字符）
- 文件上传下载（不依赖apiserver的kubectl cp）
- oom事件捕获
- kubectl debug

## Demo演示
### SSH
![2b95d08c-6154-42b8-b195-92ff0097c8d3.gif](https://i.loli.net/2021/01/19/529KgtDqbRcEB6M.gif)
### CMDB
![c84bcbdb-857e-4680-8174-5f18b160ac59.gif](https://i.loli.net/2021/01/19/wPiaLsvonOUNbzV.gif)

## 快速指南

### 1. Preparation
- kubernetes集群，可通过dockin-installer实现离线安装，**dockin-installer： [https://github.com/WeBankFinTech/Dockin-installer](https://github.com/WeBankFinTech/Dockin-installer)**
- 提前部署dockin-rm，opserver需要调用rm接口获取信息，**dockin-rm [https://github.com/WeBankFinTech/Dockin-rm](https://github.com/WeBankFinTech/Dockin-rm)**
- 准备redis，redis中存放了一些shell命令的黑白名单，apiserver通过informer推送的pod变更信息，可通过以下命令快速运行redis：
```
docker run -p 6379:6379 -d redis:latest redis-server
```
- 规划部署opserver的服务器，记录ip

