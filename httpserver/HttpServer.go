package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	r := &Frame{}
	r.addFilter(logFilter) //添加过滤器

	r.addRoute("/healthz", health)
	r.start()
}

// 自定义Filter
type Filter func(f http.HandlerFunc) http.HandlerFunc

// 存储所有自定义filter
type Frame struct {
	filters []Filter
}

//添加filter
func (f *Frame) addFilter(filter Filter) {
	f.filters = append(f.filters, filter)
}

//添加路由
func (r *Frame) addRoute(pattern string, f http.HandlerFunc) {
	r.process(pattern, f, len(r.filters)-1)
}

//绑定handler
func (r *Frame) process(pattern string, f http.HandlerFunc, index int) {
	if index == -1 {
		http.HandleFunc(pattern, f)
		return
	}

	fWrap := r.filters[index](f)
	index--
	r.process(pattern, fWrap, index)
}

//启动http-server
func (r *Frame) start() {
	if err := http.ListenAndServe(":80", nil); err != nil {
		panic(err)
	}
}

//日志过滤器
func logFilter(f http.HandlerFunc) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {
		ip := getIp(request)
		start := time.Now()
		currentTime := time.Now().Format("2006-01-02 15:04:05")

		f.ServeHTTP(writer, request)
		fmt.Printf("==== reqtime:%s; req url:%s; result:%s; remote ip:%s; time cost:%s ====\n", currentTime, ip, "200", time.Since(start))
	}
}

// 获取客户端真实ip地址
func getIp(r *http.Request) string {

	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip
	}

	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	if net.ParseIP(ip) != nil {
		return ip
	}

	return ""
}

// 健康检查
func health(wr http.ResponseWriter, request *http.Request) {

	for k, v := range request.Header {
		wr.Header().Set(k, strings.Join(v, ""))
		fmt.Printf("k%s=v%s\n", k, v)
	}

	version := os.Getenv("VERSION")
	wr.Header().Set("VERSION", version)

	io.WriteString(wr, "ok\n")
}
