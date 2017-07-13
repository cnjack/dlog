# dlog
aliyun simple log(带缓存写出) &amp;&amp; os.stdout写入

# Usage

```go
package main

import (
	"github.com/sirupsen/logrus"
	"github.com/cnjack/dlog"
	"time"
)

var writer *dlog.Writer

func main(){
	var log = logrus.New()


	var err error
	writer,err = dlog.NewWriter("xxxxxx.cn-hangzhou.log.aliyuncs.com", "xxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxx","xxxxxx", "xxxxxx")
	if err != nil {
		panic(err)
	}
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.Out = writer
	log.Formatter = &logrus.JSONFormatter{}

	log.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")
	writer.DoWrite()
	time.Sleep(time.Second*10)
}
```