### 1、list
```
opsctl list resource resourcename [subsysName|subsysId|ip] -d [dcn] 
opsctl list resource dockin-test ##获取子系统dockin-test下pods相关信息
```

### 2、get
```
opsctl get pods
opsctl get nodes
opsctl get pods dockin-test-20190321-152445825
opsctl get node 192-168-1-101
```
### 3、exec
```
opsctl exec pod command args

exec lss
opsctl exec dockin-test-20190321-152445825 less /data/cluster/dockin.yaml

exec vi
opsctl exec dockin-test-20190321-152445825 vi /data/cluster/dockin.yaml
```
