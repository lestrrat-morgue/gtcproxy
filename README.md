gtcproxy
========

An exercise in "What If I wrote tcproxy in Go?"

```
go get github.com/lestrrat/gtcproxy
gtcproxy "11211 -> 11212" // proxy port 11211 to 11212
gtcproxy "127.0.0.1:11211 127.0.0.1:11212" // same, with explicit IP
```
