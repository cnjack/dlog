package dlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"context"

	"github.com/cnjack/dlog/pb"
	"github.com/golang/protobuf/proto"
	"os"
)

const (
	LOG_NUM  = 200               // 当日志条数达到 200条时 触发 写入ali log 请求
	LOG_SIZE = 2.5 * 1024 * 1024 //当日志大小达到2.5M时 触发 写入 ali log 请求
)

var sigs = make(chan int, 1)

type Writer struct {
	url          string
	accessKey    string
	accessSecret string
	logStore     string
	log          *pb.LogGroup
	client       LogClient
	lock         sync.Mutex
}

func NewWriter(url, accessKey, accessSecret, logStore, topic string, ctx context.Context) (w *Writer, err error) {
	w = &Writer{
		url:          url,
		accessKey:    accessKey,
		accessSecret: accessSecret,
		logStore:     logStore,
		log: &pb.LogGroup{
			Topic: &topic,
		},
	}
	w.client, err = NewAliLogClient(w.url, w.accessKey, w.accessSecret)
	if err != nil {
		return nil, err
	}
	go func() {
		//保证没10分钟写一次数据
		ticker := time.NewTicker(time.Duration(600) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if len(w.log.Logs) > 0 {
					w.DoWrite()
				}
			case <-ctx.Done():
				if len(w.log.Logs) > 0 {
					w.DoWrite()
				}
			}
		}
	}()
	return w, nil
}

func (w *Writer) SetClient(client LogClient) {
	w.client = client
}

func (w *Writer) Write(log []byte) (int, error) {
	w.lock.Lock()
	newLog := &pb.Log{
		Time: proto.Uint32(uint32(time.Now().Unix())),
	}
	var logDataInf interface{}
	err := json.Unmarshal(log, &logDataInf)
	if err != nil {
		w.lock.Unlock()
		return 0, err
	}
	if logData, ok := logDataInf.(map[string]interface{}); ok {
		for k, v := range logData {
			content := &pb.Log_Content{
				Key:   proto.String(k),
				Value: proto.String(fmt.Sprint(v)),
			}
			newLog.Contents = append(newLog.Contents, content)
		}
	} else {
		w.lock.Unlock()
		return 0, errors.New("log is not json map[string]string")
	}
	w.log.Logs = append(w.log.Logs, newLog)
	aliLogBytes, _ := proto.Marshal(w.log)

	//ali_log 官方文档: 日志一次写入条数超过4096条 或大小超过3M, 超过则写入失败
	if len(w.log.Logs)+1 >= LOG_NUM || len(aliLogBytes) > LOG_SIZE {
		w.lock.Unlock()
		w.DoWrite()
		return 0, nil
	}

	w.lock.Unlock()
	return os.Stdout.Write(log)
}

func (w *Writer) DoWrite() {
	w.lock.Lock()
	n := copyAndEmpty(w.log)
	w.lock.Unlock()

	logdata, _ := proto.Marshal(n)
	sendLog(w, &logdata, 0)
	return
}

func sendLog(w *Writer, logdata *[]byte, times int) {
	if times > 2 {
		//3次之后就不再尝试
		return
	}
	time.Sleep(time.Duration(2*times) * time.Second)
	_, err := w.client.Send("POST", nil, *logdata, fmt.Sprintf("logstores/%s", w.logStore))
	times += 1
	if err != nil {
		sendLog(w, logdata, times)
		return
	}
}

func copyAndEmpty(l *pb.LogGroup) *pb.LogGroup {
	n := &pb.LogGroup{
		Topic: l.Topic,
		Logs:  l.Logs,
	}
	l.Logs = nil
	return n
}
