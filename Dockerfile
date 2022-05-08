FROM golang:1.6-alpine as build
ADD . /
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o HttpServer .

# 运行：使用scratch作为基础镜像
FROM scratch as prod

# 在build阶段复制可执行的go二进制文件app
COPY --from=build /app/HttpServer /

#启动服务
CMD ["/HttpServer"]

EXPOSE 80