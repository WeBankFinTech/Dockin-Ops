## dockin-opagent
The agent of dockin is deployed in each kubernetes node through daemonset, and it mainly has the following functions
- Mount docker.sock to connect to dockerd
- Integrate docker api for docker exec operation
- Manage the correspondence between containerId and podName in the current node
- Provide spdy interface to respond to exec requests initiated by dockin-opserver
- Bind the input and output streams of docker api to the input and output streams of spdy

### Configuration Management
Modify the configuration file application.yaml, you need to pay attention to the access address of rm
```yaml
app:
  rm:
    api: http://127.0.0.1:10002/rmController # RM access address
  container:
    ticker: 30 # The update frequency of the container in the current node
  http:
    port: 8085 # Provided port address
  debug:
    port: 10102 # goprof address
  docker:
    sock: unix:///var/run/docker.sock # docker.sock directory
  qos:
    path: /data/cgroup # cgroup mount directory
  logs:
    cmd-white-list: # The list of commands supported by the logs command
      -grep
      -zgrep
      -cat
      -head
      -tail
      -awk
      -uniq
      -sort
      -ls
    cmd-timeout: 5000 # docker command execution timeout
    max-file-size: 3000 # File operation size limit
    max-line: 1000 # The maximum number of lines for tail and other commands
    root: /data/logs/ # Mounted host log directory
```

### daemonset yaml file
The most important configuration:
> hostIPC: true

Without this configuration, docker.sock cannot be mounted, so the execution of docker api will fail

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
        -image: REGISTRY_ENDPOINT/dockin-opagent:0.1.0
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
            -mountPath: /data/logs/
              name: logs-dir
              readOnly: false
            -mountPath: /var/run/
              name: docker-run
              readOnly: false
            -mountPath: /data/cgroup/
              name: kubepods
              readOnly: false
            -mountPath: /logs/
              name: log
              readOnly: false
            -mountPath: /data/dockin_prestop_info/
              name: prestop-info
              readOnly: false
      hostNetwork: true
      restartPolicy: Always
      terminationGracePeriodSeconds: 30
      volumes:
        -hostPath:
            path: /data/logs/
            type: DirectoryOrCreate
          name: logs-dir
        -hostPath:
            path: /var/run/
            type: DirectoryOrCreate
          name: docker-run
        -hostPath:
            path: /sys/fs/cgroup/
            type: DirectoryOrCreate
          name: kubepods
        -hostPath:
            path: /data/logs/dockin-opagent
            type: DirectoryOrCreate
          name: log
        -hostPath:
            path: /data/dockin_prestop_info
            type: DirectoryOrCreate
          name: prestop-info
```

## Compile image
The Makefile file is provided in the project, and you can directly execute ```docker-build```

## Run
```shell
cd /pathtoproject/bin
sh start.sh
```