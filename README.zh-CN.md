# Dockin Ops - Dockin Operation service

English | [中文](README.zh-CN.md)

Dockin运维管理系统是安全的运维管理服务，优化exec执行性能，支持命令权限管理


![Architecture](docs/images/dockin.png)

## 快速指南

### 1. Preparation
- k8s集群
- 提前部署dockin rm，opserver需要调用rm接口获取信息
- 准备redis，可通过以下命令快速运行redis：
```
docker run -p 6379:6379 -d redis:latest redis-server
```
- 规划部署opserver的服务器，记录ip

### 2. Compile

#### 2.1 Dockin-opserver
- 修改配置文件application.yaml，需要注意的主要为rm的地址
```
rm-address: http://127.0.0.1:10002/rmController     # RM访问地址
batch-timeout: 5000
http-port: 8084                                     # opserver的监听端口
cmd-filter-type: blacklist
while-list-update-time: 60000
limits:
  exec-forbidden:
    - vi
  file-max-size: 1000
  upload-file-max-size: 500
  download-file-max-size: 4000
  vi-file-max-size: 10
  k8s-qos: 40
  k8s-burst: 60
opagent-port: 8085                                  # opagent的监听端口
redis:
  expiration: 120000
accounts:                                           # opserver的用户信息，当前在配置文件中配置
  - account:
      user-name: app
      passwd: passwd
```
- 编译：执行以下命令即可
```
make 
```

#### 2.2 Dockin-opsctl
- 修改opserver访问地址
```
# 需要修改的文件：internal/common/url.go，将常量RemoteHost改为opserver对应的ip和端口
const RemoteHost = "127.0.0.1:8084"
```
- 编译：执行make命令
```
make
```


#### 2.3 Dockin-opagent
- 修改配置文件application.yaml，需要注意的为rm的访问地址
```
app:
  rm:
    api: http://127.0.0.1:10002/rmController  # RM访问地址
  container:
    ticker: 30
  http:
    port: 8085
  debug:
    port: 10102
  ims:
    logroot: /data/logs/
  docker:
    sock: unix:///var/run/docker.sock
  qos:
    path: /data/cgroup
  logs:
    cmd-white-list:
      - grep
      - zgrep
      - cat
      - head
      - tail
      - awk
      - uniq
      - sort
      - ls
    cmd-timeout: 5000
    max-file-size: 3000
    max-line: 1000
    root: /data/logs/

```
- 编译并打包opagent到docker镜像
```
make docker-build
```


### 3. Installation And Running

#### 3.1 dockin-opagent
1. 导出需要管理的k8s集群的配置文件，放置在configs/cluster目录下，并在原始配置文件的基础上增加dockin段，示例如下所示，需要关注的请看对应备注：
```
apiVersion: v1
clusters:
- cluster:                          # 集群的访问地址及名字，可声明多个
    insecure-skip-tls-verify: true
    server: https://127.0.0.1:6443
  name: kubernetes
contexts:                           # 上下文信息，主要用于与上述集群信息对应，设置部分配置
- context:                          
    cluster: kubernetes             # 集群名，与cluster段中的集群名对应
    namespace: test                  # 使用该配置操作的命名空间
    user: kubernetes-readonly-user  # 访问该集群所使用的用户
  name: readonly-user
current-context: readonly-user      # 默认使用的上下文
kind: Config
preferences: {}
users:
- name: kubernetes-readonly-user    # 用户信息，与context段中的用户对应
  user:
    password: your_password         # 用户密码
    username: readonly-user     # 用户名
dockin:                             # 额外自定义配置，用户声明该集群适用于的规则及对应的集群id，并声明默认的白名单
  cluster-id: ft01
  rule: test                         
  whitelist:
    - 127.0.0.1
```
2. 将项目中的*start.sh*、*configs*目录以及编译出来的可执行文件上传至服务器，执行以下命令：
```
sh start.sh
```

#### 3.2 dockin-opsctl
将可执行文件拷贝至服务器即可使用，使用如下命令查看帮助：
```
dockin-opsctl -h
```

#### 3.3 dockin-opagent
opagent以daemonSet的方式运行在k8s集群中，可直接参照项目internal/docs目录中的daemonSet样例，修改对应的镜像信息后直接应用到k8s集群即可
