gtcproxy
========

An exercise in "What If I wrote tcproxy in Go?"

## DESCRIPTION

Currently this is just "an exercise". If you want to use it for
something real, you might want to first drop me a line so
I take the time and care to harden the program!

## BINARY BUILDS (RECOMMENDED)

See the [releases page](https://github.com/lestrrat/gtcproxy/releases)

## MANUAL INSTALL

```
go get github.com/lestrrat/gtcproxy
```

## USAGE

```
gtcproxy "11211 -> 11212" // proxy port 11211 to 11212
gtcproxy "127.0.0.1:11211 127.0.0.1:11212" // same, with explicit IP
```
