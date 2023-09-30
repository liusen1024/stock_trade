FROM golang as golang

# 配置模块代理
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

ADD . /www

# 进入工作目录
WORKDIR /www

# 打包 AMD64 架构
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o go_server

FROM scratch

# 暴露服务端口
EXPOSE 9090

WORKDIR /www

# 复制打包的 Go 文件到系统用户可执行程序目录下
COPY --from=golang /www/go_server /www

# 容器启动时运行的命令
ENTRYPOINT ["/www/go_server"]