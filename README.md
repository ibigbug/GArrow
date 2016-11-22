GArrow
------

A HTTP/HTTPS ~~proxy~~ **relay** over TCP.

It works somehow like the way shadowsocks working, however GArrow provides HTTP layer proxy while shadowsocks provides socks5 proxy.

This projects is created for demostration of how HTTP related stuffs works and leanring the Golang `net/http` library.

The default encryption method used by GArrow is *aes-256-cfb* and more type of ciphers are still need to be supported.

## Installation

```
$ go get github.com/ibigbug/GArrow/cmd/garrow
```

## Configuration

```
$ cat g-arrow.yml

server: '0.0.0.0:9999'
local: '127.0.0.1:9998'
password: 'abc'
```

## Usage

### Server

```
$ garrow -m server
```

### Client

```
$ garrow -m client
```

## TODO

* [x] make better log
* [x] encryption
* [ ] distribution packaging
* [ ] documentation
* [ ] relay other protocol like pure HTTP/S proxy, shadowsocks protocol etc.
* [ ] ~~https connection reuse~~ It's impossible, since we don't know the data trasmitting and the point to release the connection.
