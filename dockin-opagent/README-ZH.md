## dockin-opagent
dockin的agent，通过daemonset的方式部署在各个kubernetes节点中，主要有如下功能
- 挂载docker.sock，用于连接dockerd
- 集成docker api，进行docker exec操作
- 管理当前节点中containerId和podName的对应关系
- 提供spdy接口，用于响应dockin-opserver发起的exec请求
- 绑定docker api的输入输出流到spdy的输入输出流

### 配置管理
修改配置文件application.yaml，需要注意的为rm的访问地址
```yaml
app:
  rm:
    api: http://127.0.0.1:10002/rmController  # RM访问地址
  container:
    ticker: 30                                  # 当前节点中container更新频率
  http:
    port: 8085                                  # 提供的端口地址
  debug:
    port: 10102                                 # goprof地址
  docker:
    sock: unix:///var/run/docker.sock           # docker.sock目录
  qos:
    path: /data/cgroup                          # cgroup挂载目录
  logs:
    cmd-white-list:                             # logs命令支持的命令列表
      - grep
      - zgrep
      - cat
      - head
      - tail
      - awk
      - uniq
      - sort
      - ls
    cmd-timeout: 5000                           # docker命令执行timeout
    max-file-size: 3000                         # 文件操作大小限制
    max-line: 1000                              # tail 等命令最大行数限制
    root: /data/logs/                           # 挂载的宿主机日志目录
```

### daemonset yaml文件
其中最重要的配置：
> hostIPC: true
若没有这项配置，则docker.sock是无法挂载上来，从而执行docker api会失败

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: dockin-opagent
  name: dockin-opagent
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: dockin-opagent
  template:
    metadata:
      labels:
        k8s-app: dockin-opagent
    spec:
      hostIPC: true
      hostPID: true
      containers:
        - image: REGISTRY_ENDPOINT/dockin-opagent:0.1.0
          imagePullPolicy: Always
          name: dockin-opagent
          resources:
            limits:
              cpu: "2000m"
              memory: "500Mi"
            requests:
              cpu: "100m"
              memory: "200Mi"
          securityContext:
            privileged: false
          volumeMounts:
            - mountPath: /data/logs/
              name: logs-dir
              readOnly: false
            - mountPath: /var/run/
              name: docker-run
              readOnly: false
            - mountPath: /data/cgroup/
              name: kubepods
              readOnly: false
            - mountPath: /logs/
              name: log
              readOnly: false
            - mountPath: /data/dockin_prestop_info/
              name: prestop-info
              readOnly: false
      hostNetwork: true
      restartPolicy: Always
      terminationGracePeriodSeconds: 30
      volumes:
        - hostPath:
            path: /data/logs/
            type: DirectoryOrCreate
          name: logs-dir
        - hostPath:
            path: /var/run/
            type: DirectoryOrCreate
          name: docker-run
        - hostPath:
            path: /sys/fs/cgroup/
            type: DirectoryOrCreate
          name: kubepods
        - hostPath:
            path: /data/logs/dockin-opagent
            type: DirectoryOrCreate
          name: log
        - hostPath:
            path: /data/dockin_prestop_info
            type: DirectoryOrCreate
          name: prestop-info
```

## 编译image
在项目中提供了Makefile文件，直接执行```docker-build```即可

## 运行
```shell
cd /pathtoproject/bin
sh start.sh
```