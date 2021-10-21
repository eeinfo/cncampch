#### 构建本地镜像。
##### 编写 Dockerfile 将练习 2.2 编写的 httpserver 容器化（请思考有哪些最佳实践可以引入到 Dockerfile 中来）。

- 使用多阶段构建，大幅减少镜像体积。
- 不要安装不必要的包，安装包使用较快的源。
- 对多行参数进行排序。
- 设置健康检查。
- 将日志文件保存到宿主机，避免因日志文件导致容器过大。


```
# 构建镜像
$ sudo docker build -t eeinfo/http-server:v1.0 .

...
Successfully built 0b......
Successfully tagged eeinfo/http-server:v1.0

# 测试镜像
$ sudo docker run -it eeinfo/http-server:v1.0 -p 80:80

[GIN-debug] GET    /healthz                  --> main.healthzHandler (6 handlers)
[GIN-debug] GET    /error                    --> main.errorHandler (6 handlers)
[GIN-debug] GET    /test                     --> main.testHandler (6 handlers)

```

#### 将镜像推送至 Docker 官方镜像仓库。
```
# 根据提示输入docker hub用户名和密码
$ sudo docker login
...
Username: ...
Password: 

# 推送镜像
$ sudo docker push eeinfo/http-server:v1.0
The push refers to repository [docker.io/eeinfo/http-server]
0b......: Pushed 
4b......: Pushed 
e2......: Mounted from library/alpine 
v1.0: digest: sha256:da...... size: 950
```

#### 通过 Docker 命令本地启动 httpserver。

```
$ sudo docker run -it eeinfo/http-server:v1.0 -p 80:80

Unable to find image 'eeinfo/http-server:v1.0' locally
v1.0: Pulling from eeinfo/http-server
...
[GIN-debug] GET    /healthz                  --> main.healthzHandler (6 handlers)
[GIN-debug] GET    /error                    --> main.errorHandler (6 handlers)
[GIN-debug] GET    /test                     --> main.testHandler (6 handlers)
```

#### 通过 nsenter 进入容器查看 IP 配置。
```
$ sudo docker container ls
# 省略部分
CONTAINER ID   IMAGE                                               
c8124f127827   eeinfo/http-server:v1.

# 得到PID
$ sudo docker inspect --format {{.State.Pid}} c8124f127827
434388

# 使用nsenter命令进入容器，查看 IP 配置
$ sudo nsenter -t 434388 -n ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
12: eth0@if13: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
```