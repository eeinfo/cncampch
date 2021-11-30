**优雅启动**

通过在`startupProbe`中调用`/healthz`进行验证是否启动成功，如果状态为`Success`则启动成功。再通过`readinessProbe`验证普通接口是否可以调用成功，如果状态为`Success`则开始接入流量。

```yaml
# Deployment yaml
startupProbe:
  httpGet:
    path: /healthz
    port: 80
  initialDelaySeconds: 10
  periodSeconds: 5
readinessProbe:
  httpGet:
    path: /test
    port: 80
  initialDelaySeconds: 3
  periodSeconds: 5
  failureThreshold: 5
```

**优雅终止**

在程序接收`SIGTERM`信号来优雅终止程序，等待60秒。因为没有设置`preStop`，只需要将`terminationGracePeriodSeconds`设置为70秒，防程序繁忙处理信号延迟，这里设置的比程序中等待时间略长。

```go
// 实现优雅关闭, 尽量等待请求完成，如果超过60秒请求未完成，则退出
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
log.Println("Shutting down server...")
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
if err := server.Shutdown(ctx); err != nil {
	log.Fatal("Server forced to shutdown:", err)
}
log.Println("Server exiting")
```

```yaml
terminationGracePeriodSeconds: 70
```

**资源需求和 `QoS` 保证**

设置`requests` 和 `limits` 比例为 1:2，`QoS` 级别为 `Burstable`，让Pod可以更容易找到节点运行，并通过`PriorityClass`控制Pod为高优先级，同时抢占策略为不抢占其他Pod。

```yaml
# Deployment yaml
containers:
- name: http-server
  image: eeinfo/http-server:v1.0
  imagePullPolicy: IfNotPresent   
  resources:
    limits:
      memory: 1Gi
      cpu: 1
    requests:
      memory: 512Mi
      cpu: 500m
...
priorityClassName: hige-priority
```

```yaml
# PriorityClass yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000000
globalDefault: false
preemptionPolicy: Never
description: "high priority Pod."
```

**探活**

通过在`livenessProbe`中调用`/healthz`进行探活，超时时间为3秒，失败5次后kubelet会杀死容器并重新启动容器。

```yaml
# Deployment yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 80
  initialDelaySeconds: 3
  periodSeconds: 5
  timeoutSeconds: 3
  successThreshold: 1
  failureThreshold: 5
```

**日常运维需求，日志等级**

将日志等级设置为异常级别，减少日志输出。

```go
//设置日志级别
logger.SetLevel(logrus.ErrorLevel)
```

**配置和代码分离**

`ConfigMap` 将您的环境配置信息和 容器镜像 解耦，便于应用配置的修改。

```yaml
# ConfigMap yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: env-config
  namespace: default
data:
  version: v2.0
```

```yaml
# Deployment yaml
env:
- name: VERSION
  valueFrom:
    configMapKeyRef:
      name: env-config
      key: version
```

```bash
# 创建ConfigMap
$ k create -f config-map.yaml 

# 创建PriorityClass
$ k create -f high-priority.yaml 

# 创建Deployment
$ k create -f http-server.yaml 

# 查看Pod状态
$ kubectl get pod --namespace default
NAME                          READY   STATUS    RESTARTS   AGE
http-server-9d7576b4b-76k2c   1/1     Running   0          92m
http-server-9d7576b4b-dtnpm   1/1     Running   0          92m
http-server-9d7576b4b-qqf8n   1/1     Running   0          92m

```

