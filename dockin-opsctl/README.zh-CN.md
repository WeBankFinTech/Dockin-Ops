## dockin-opsctl
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

[English](README.md) | Chinese

dockin-opsctl是一款类似kubectl客户端二进制客户端，基于cobra开发的客户端，用于和docker-opserver建立连接，具有如下功能:
- 绑定当前输入输出
- raw模式管理
- 建立与dockin-opserver的http和websocket连接


### 编译
修改opserver访问地址
```GO
# 需要修改的文件：internal/common/url.go，将常量RemoteHost改为dockin-opserver对应的ip和端口
const RemoteHost = "127.0.0.1:8084"
```

执行项目中的Makefile即可
```shell
make
```

### 运行
可以通过./dockin-opsctl -h查看当前工具的使用方式

```shell
dockin-opsctl used to execute cmd in dockin-opserver

Usage:
  dockin-opsctl [flags]
  dockin-opsctl [command]

Available Commands:
  auth        auth
  exec        exec cmd in pod
  get         Display one or many resources
  help        Help about any command
  list        get resource info from rm interface
  ssh         ssh to pod

Flags:
  -h, --help                    help for dockin-opsctl
  -n, --namespace string        If present, the namespace scope for this CLI request
      --profile string          Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex) (default "none")
      --profile-output string   Name of the file to write the profile to (default "profile.pprof")
  -r, --rule string
  -v, --version                 show version and exit

Use "dockin-opsctl [command] --help" for more information about a command.
```

可以通过./dockin-opsctl ssh help查看子命令的说明
```
./dockin-opsctl ssh -h
                # ssh to pod, according the the podName, and password, userName
                # podName is need, userName or password will alter to input if not provided
                # only allowed to ssh pods which you owns the access authority,
                # for instance, developers A belongs to department tctp, he can not
                # ssh to pods belongs to department tdtp
                #
                # for instance
                # ssh according to podName
                #       dockin-opsctl ssh dockin-test-20191012-182050448-0 -u admin -p admin -r default
                # ssh according to pod ip
                # dockin-opsctl ssh 192.168.1.1 -u admin -p admin -r default
                # ssh according to access token
                # dockin-opsctl ssh 192.168.1.1 --access-token foiudepjfpghuqwipr1028390eu8fihyedpqrhfuwospkal

Usage:
  dockin-opsctl ssh to pod

Examples:
dockin-opsctl ssh dockin-test-20191012-182050448-0 -u admin -p admin -r default

Flags:
  -a, --access-token string   access token to access the pod by ssh command, generate from auth command
  -c, --container string      ContainerName name. If omitted, the first container in the pod will be chosen (default "op_server")
  -e, --env stringArray       exec env variable, like: a=1
  -h, --help                  help for ssh
  -p, --password string       pass password
  -s, --user string           run as user name, default to app
  -u, --user name string      pass user name
  -w, --work-dir string       exec work directory

Global Flags:
  -n, --namespace string        If present, the namespace scope for this CLI request
      --profile string          Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex) (default "none")
      --profile-output string   Name of the file to write the profile to (default "profile.pprof")
  -r, --rule string
  -v, --version                 show version and exit
```