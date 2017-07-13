package main

import (
	"github.com/sirupsen/logrus"
	"github.com/cnjack/dlog"
	"time"
	"context"
)

func main(){
	var log = logrus.New()
	ctx, done := context.WithCancel(context.Background())
	writer,err := dlog.NewWriter("xxxxxx.cn-hangzhou.log.aliyuncs.com", "xxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxx","xxxxxx", "xxxxxx", ctx)
	if err != nil {
		panic(err)
	}

	log.Out = writer
	log.Formatter = &logrus.JSONFormatter{}

	log.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")
	//make sure write
	done()
	time.Sleep(time.Second*5)
}