# vt-go

[English](./README.MD) | 简体中文

## 简介

一个用于解析终端输入输出的工具。

## 使用方式

完整代码可查看 [example](./examples/main.go) 。

```go

package main

import (
	"log"

	"github.com/dushixiang/vt"
)

func main() {
	content, err := readInputContent()
	if err != nil {
		log.Fatal(err)
	}

	v := vt.New()
	v.Advance(content)
	lines := v.Result()
	for _, line := range lines {
		println(line)
	}
}
```

输出

```bash
> # Welcome to asciinema!
> # See how easy it is to record a terminal session
> # First install the asciinema recorder
> brew install asciinema
==> Downloading https://homebrew.bintray.com/bottles/asciinema-2.0.2_2.catalina.bottle.1.tar.gz
==> Downloading from https://akamai.bintray.com/4a/4ac59de631594cea60621b45d85214e39a90a0ba8ddf4eeec5cba34bd6145711
######################################################################## 100.0%
==> Pouring asciinema-2.0.2_2.catalina.bottle.1.tar.gz
🍺  /usr/local/Cellar/asciinema/2.0.2_2: 613 files, 6.4MB
> # Now start recording
> asciinema rec
asciinema: recording asciicast to /tmp/u52erylk-ascii.cast
asciinema: press <ctrl-d> or type "exit" when you're done
bash-3.2$ # I am in a new shell instance which is being recorded now
bash-3.2$ sha1sum /etc/f* | tail -n 10 | lolcat -F 0.3
da39a3ee5e6b4b0d3255bfef95601890afd80709  /etc/find.codes
88dd3ea7ffcbb910fbd1d921811817d935310b34  /etc/fstab.hd
442a5bc4174a8f4d6ef8d5ae5da9251ebb6ab455  /etc/ftpd.conf
442a5bc4174a8f4d6ef8d5ae5da9251ebb6ab455  /etc/ftpd.conf.default
d3e5fb0c582645e60f8a13802be0c909a3f9e4d7  /etc/ftpusers
bash-3.2$ # To finish recording just exit the shell
bash-3.2$ exit
exit
asciinema: recording finished
asciinema: press <enter> to upload to asciinema.org, <ctrl-c> to save locally

https://asciinema.org/a/17648
> # Open the above URL to view the recording
> # Now install asciinema and start recording your own sessions
> # Oh, and you can copy-paste from here
> # Bye!
```