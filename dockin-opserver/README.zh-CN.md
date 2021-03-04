# Dockin-opserver
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

[English](README.md) | Chinese

dockin-opserver是基于kubernetes client-go开发的apiserver接口代理，支持如下基本功能：
- 多集群管理，可以同时管理多套apiserver集群
- 用户账号信息管理，Pod本身访问时没有账号管理的，因此只要具有kubeconfig就可以访问所有Pod，对跨部门等情况来说，存在安全隐患
- ssh代理，用户通过kubectl exec /bin/bash -it的方式登录到Pod中之后，可以操作任何命令，但是由于kubernetes的特性，对memory存在严格的控制，因此一些耗费memory的操作，比如vi一个特大文件，就很容易造成Pod被OOM kill，通过ssh代理，我们可以拦截所有用户执行的命令，并且进行安全判断（黑白名单）。
- 审计，通过kubectl exec的方式可以执行任何命令，那如何用户执行了何种命令而导致了安全隐患，通过审计功能，我们将用户执行的所有命令，无论是exec还是在shell环境中，都有相对应的存档，做到有迹可循。
- 提供http和websocket接口，用于执行普通exec和交互式exec的请求
- 协议转换，将websocket和spdy协议数据互相转换
- 通过client-go的informer功能，将pod的Add、Update、Delete的时间保存到redis中

### 配置介绍
```yaml
rm-address: http://127.0.0.1:10002/rmController     # RM访问地址
batch-timeout: 5000
http-port: 8084                                     # opserver的监听端口
cmd-filter-type: blacklist                          # 命令拦截模式，blacklist(黑名单)、whitelist(白名单)
while-list-update-time: 60000                       # 黑白名单更新频率
limits:                                             # 限制
  exec-forbidden:                                   # exec命令限制的
    - vi                                            
  file-max-size: 1000                               # 可操作性文件的最大大小，单位M
  upload-file-max-size: 500                         # 文件上传的最大大小，单位M
  download-file-max-size: 4000                      # 文件下载的最大大小，单位M
  vi-file-max-size: 10                              # vi可操作性的文件最大大小
opagent-port: 8085                                  # opagent端口
redis:
  expiration: 120000                                # redis key失效时间
accounts:                                           # opserver的用户信息，当前在配置文件中配置
  - account:
      user-name: app
      passwd: passwd
```

### kubeconfig管理
导出需要管理的k8s集群的配置文件，放置在configs/cluster目录下，并在原始配置文件的基础上增加dockin段，示例如下所示，需要关注的请看对应备注：
```yaml
apiVersion: v1
clusters:
- cluster:                          # 集群的访问地址及名字，可声明多个
    insecure-skip-tls-verify: true
    server: https://127.0.0.1:6443
  name: kubernetes
contexts:                           # 上下文信息，主要用于与上述集群信息对应，设置部分配置
- context:                          
    cluster: kubernetes             # 集群名，与cluster段中的集群名对应
    namespace: test                 # 使用该配置操作的命名空间
    user: kubernetes-readonly-user  # 访问该集群所使用的用户
  name: readonly-user
current-context: readonly-user      # 默认使用的上下文
kind: Config
preferences: {}
users:
- name: kubernetes-readonly-user    # 用户信息，与context段中的用户对应
  user:
    password: your_password         # 用户密码
    username: readonly-user         # 用户名
dockin:                             # 额外自定义配置，用户声明该集群适用于的规则及对应的集群id，并声明默认的白名单
  cluster-id: ft01                  # 集群id
  rule: test                        # 规则组，若需要针对一份证书管理不同的白名单，则可以通过rule扩展
  whitelist:                        
    - 127.0.0.1                     # 可允许执行的ip白名单，当前证书对应的集群，只允许这些ip访问
```

### 编译
我们在项目中提供了Makefile，可以直接通过make进行编译，会生成相对应的tar包

### 运行
```
cd /pathtoproject/bin
sh start.sh
```