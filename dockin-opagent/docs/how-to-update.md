how to remove ds
---
1. remove ds
```
kubectl delete ds dockin-opagent -n kube-system
```
2. check pod exist
```
kubectl get pods -n kube-system|grep dockin-opagent
```

how to install ds
---
1. check all pod removed
```
kubectl get pods -n kube-system|grep dockin-opagent
```
2. install new ds
```
kubectl apply -f daemonset.yaml -n kube-system 
```
3. check pod installed
```
kubectl get pods -n kube-system|grep dockin-opagent
```