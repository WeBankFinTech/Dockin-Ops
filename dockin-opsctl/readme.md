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

### 4、file
```
opsctl file [podName] [-t upload/download] [-s srcDir] [-d dstDir]
```

####4.1 目录上传
opsctl file dockin-test-20190321-152445825 -t upload -s /data/v_wbzfhe -d /data

####4.2 文件上传 需指定dest
```
正确的请求
opsctl file dockin-test-20190321-152445825 -t upload -s /data/test.log -d /data/test.log
错误的请求    
opsctl file dockin-test-20190321-152445825 -t upload -s /data/test.log -d /data
```   

####4.3 目录下载
```
opsctl file dockin-test-20190321-152445825 -t download -s /data/cluster -d /data
```

####4.4 文件下载
```
opsctl file dockin-test-20190321-152445825 -t download -s /data/cluster/dockin.yaml -d /data
```

####5、jvm
```
opsctl jvm [podName] [jstat/jps/jmap] [arglist] 
opsctl jvm dockin-test-20190321-152445825 jps
```