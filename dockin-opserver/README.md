# Dockin-opserver
dockin-opserver is an apiserver interface agent developed based on kubernetes client-go, which supports the following basic functions:
- Multi-cluster management, can manage multiple sets of apiserver clusters at the same time
- User account information management. There is no account management when the Pod itself is accessed. Therefore, as long as you have kubeconfig, you can access all Pods. For cross-departmental situations, there are security risks
- ssh proxy, after logging in to Pod through kubectl exec /bin/bash -it, users can operate any command, but due to the characteristics of kubernetes, there is strict control over memory, so some memory-consuming operations, such as vi Very large files can easily cause Pod to be OOM kill. Through ssh proxy, we can intercept all commands executed by users and perform security judgments (black and white lists).
- Audit. Any command can be executed through kubectl exec. How can the user execute which command causes a security risk. Through the audit function, we will check all the commands executed by the user, whether it is exec or in the shell environment. Corresponding archives can be traced.
- Provide http and websocket interfaces for executing ordinary exec and interactive exec requests
- Protocol conversion, convert websocket and spdy protocol data to each other
- Save the time of Add, Update, and Delete of pod to redis through the informer function of client-go

### Configuration introduction
```yaml
rm-address: http://127.0.0.1:10002/rmController # RM access address
batch-timeout: 5000
http-port: 8084 # listening port of opserver
cmd-filter-type: blacklist # Command interception mode, blacklist (blacklist), whitelist (whitelist)
while-list-update-time: 60000 # Black and white list update frequency
limits: # limits
  exec-forbidden: # exec command restricted
    -vi
  file-max-size: 1000 # The maximum size of the operable file, in M
  upload-file-max-size: 500 # The maximum size of file upload, in M
  download-file-max-size: 4000 # The maximum size of file download, in M
  vi-file-max-size: 10 # vi operable file maximum size
opagent-port: 8085 # opagent port
redis:
  expiration: 120000 # redis key expiration time
accounts: # User information of opserver, currently configured in the configuration file
  -account:
      user-name: app
      passwd: passwd
```

### kubeconfig management
Export the configuration file of the k8s cluster that needs to be managed, place it in the configs/cluster directory, and add a dockin section on the basis of the original configuration file. The example is shown below. For those who need attention, please see the corresponding notes:
```yaml
apiVersion: v1
clusters:
-cluster: # The access address and name of the cluster, you can declare multiple
    insecure-skip-tls-verify: true
    server: https://127.0.0.1:6443
  name: kubernetes
contexts: # Context information, mainly used to correspond to the above cluster information, set up some configurations
-context:
    cluster: kubernetes # cluster name, corresponding to the cluster name in the cluster section
    namespace: test # Use the namespace of the configuration operation
    user: kubernetes-readonly-user # The user used to access the cluster
  name: readonly-user
current-context: readonly-user # Context used by default
kind: Config
preferences: {}
users:
-name: kubernetes-readonly-user # User information, corresponding to the user in the context section
  user:
    password: your_password # User password
    username: readonly-user # username
dockin: # Additional custom configuration, the user declares the rules applicable to the cluster and the corresponding cluster id, and declares the default whitelist
  cluster-id: ft01 # cluster id
  rule: test # Rule group, if you need to manage different whitelists for a certificate, you can use rule extension
  whitelist:
    -127.0.0.1 # Permitted ip whitelist, the cluster corresponding to the current certificate, only these ips are allowed to access
```

### Compile
We provide Makefile in the project, which can be compiled directly by make, and the corresponding tar package will be generated

### Run
```shell
cd /pathtoproject/bin
sh start.sh
```