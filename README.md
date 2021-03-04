# Dockin Ops-Dockin Operation service

[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

English| [Chinese](README.zh-CH.md) 

Dockin operation and maintenance management system is a safe operation and maintenance management service that optimizes exec execution performance, supports command authority management, and supports scene operation and maintenance orchestration.

**For more Docking components, please visit [https://github.com/WeBankFinTech/Dockin](https://github.com/WeBankFinTech/Dockin)**

![Architecture](docs/images/dockin.png)

## Component introduction
### Dockin-Opserver
dockin-opserver is an apiserver interface agent developed based on kubernetes client-go, which supports the following basic functions:
- Multi-cluster management, can manage multiple sets of apiserver clusters at the same time
- User account information management. There is no account management when the Pod itself is accessed. Therefore, as long as you have kubeconfig, you can access all Pods. For cross-departmental situations, there are security risks
- ssh proxy, after logging in to Pod through kubectl exec /bin/bash -it, users can operate any command, but due to the characteristics of kubernetes, there is strict control over memory, so some memory-consuming operations, such as vi Very large files can easily cause Pod to be OOM kill. Through ssh proxy, we can intercept all commands executed by users and perform security judgments (black and white lists).
- Audit. Any command can be executed through kubectl exec. How can the user execute which command causes a security risk. Through the audit function, we will check all the commands executed by the user, whether it is exec or in the shell environment. Corresponding archives can be traced.
- Provide http and websocket interfaces for executing ordinary exec and interactive exec requests
- Protocol conversion, convert websocket and spdy protocol data to each other
-Save the time of Add, Update, and Delete of pod to redis through the informer function of client-go

### Dockin-Opsctl
Similar to kubectl client, binary client, user and dockin-opserver establish http or websocket request, bind current standard input and standard input, enter raw mode in interactive mode

### Dockin-Opagent
The agent of dockin is deployed in each kubernetes node through daemonset, and it mainly has the following functions
- Mount docker.sock to connect to dockerd
- Integrate docker api for docker exec operation
- Manage the correspondence between containerId and podName in the current node
- Provide spdy interface to respond to exec requests initiated by dockin-opserver
- Bind the input and output streams of docker api to the input and output streams of spdy

### dockctl
dockctl is a package of dockin-opsctl
- dockctl provides batch operation of multiple Pods, in units of subsystems
- Beautify output, dockin-opsctl returns standard output, mostly json data, dockctl will beautify the display of json data

## List of open source functions
- dockctl cmdb, limit Pod-related information based on the subsystem
- exec proxy
- ssh tool
  - Support intercepting ssh commands
  - Support command parameter interception
  - Support account management
- Pod permission management

## Roadmap
- Shell content analysis optimization (based on escape characters, control characters)
- File upload and download (kubectl cp without apiserver)
- oom event capture
- kubectl debug

## Demo
### SSH
![2b95d08c-6154-42b8-b195-92ff0097c8d3.gif](https://i.loli.net/2021/01/19/529KgtDqbRcEB6M.gif)
### CMDB
![c84bcbdb-857e-4680-8174-5f18b160ac59.gif](https://i.loli.net/2021/01/19/wPiaLsvonOUNbzV.gif)

## third-party component 
- kubernetes cluster, offline installation can be achieved through dokin-installer, **dockin-installer: [https://github.com/WeBankFinTech/Dockin-installer](https://github.com/WeBankFinTech/Dockin-installer)* *
- Deploy dokin-rm in advance, opserver needs to call the rm interface to obtain information, **dockin-rm [https://github.com/WeBankFinTech/Dockin-rm](https://github.com/WeBankFinTech/Dockin-rm) **
- Prepare redis. Redis stores a black and white list of some shell commands. The pod change information pushed by the apiserver through the informer can be quickly run through the following commands:
```
docker run -p 6379:6379 -d redis:latest redis-server
```