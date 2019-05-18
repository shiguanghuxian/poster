# 海报生成
提供http和grpc两种调用方式的海报生成服务。
可以根据公司具体情况稍加改动就可以接入服务端海报生成。

海报生成前端实现方案较多，后端实现方案可不考虑客户端兼容性问题。

## 性能
```
MacBook Pro (Retina, 15-inch, Mid 2014)

goos: darwin
goarch: amd64
pkg: github.com/shiguanghuxian/poster/program/service
BenchmarkDrawPoster                   20         100703717 ns/op        16831983 B/op     117748 allocs/op
BenchmarkDrawPoster-4                 20          83181121 ns/op        16832505 B/op     117756 allocs/op
BenchmarkDrawPoster-8                 20          83203308 ns/op        16833221 B/op     117765 allocs/op
PASS
ok      github.com/shiguanghuxian/poster/program/service        32.778s
```

## 备注
实现基本功能后实现获取微信小程序码功能

