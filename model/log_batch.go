package model

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

var (
	logQueue     chan *Log
	logQueueOnce sync.Once
	logBatchDone chan struct{}
)

func effectiveLogBatchSize() int {
	if common.LogBatchSize <= 0 {
		return 10
	}
	return common.LogBatchSize
}

func effectiveLogBatchInterval() time.Duration {
	if common.LogBatchInterval <= 0 {
		return time.Second
	}
	return time.Duration(common.LogBatchInterval) * time.Second
}

// InitLogBatchWriter starts the background goroutine that batches log inserts.
// It is a no-op when LogBatchEnabled is false.
func InitLogBatchWriter() {
	if !common.LogBatchEnabled {
		return
	}
	logQueueOnce.Do(func() {
		batchSize := effectiveLogBatchSize()
		logQueue = make(chan *Log, batchSize*10)
		logBatchDone = make(chan struct{})
		go logBatchWorker()
	})
}

// SubmitLog attempts to enqueue a log for batch insertion.
// Returns false if the queue is nil (batch writer not started) or full,
// signaling the caller to fall back to a synchronous write.
func SubmitLog(log *Log) bool {
	if logQueue == nil {
		return false
	}
	select {
	case logQueue <- log:
		return true
	default:
		return false
	}
}

// FlushLogQueue drains and writes all pending logs, then returns.
// Call this during graceful shutdown.
func FlushLogQueue() {
	if logQueue == nil {
		return
	}
	close(logQueue)
	<-logBatchDone
}

func logBatchWorker() {
	defer close(logBatchDone)
	batchSize := effectiveLogBatchSize()
	buf := make([]*Log, 0, batchSize)
	ticker := time.NewTicker(effectiveLogBatchInterval())
	defer ticker.Stop()

	flush := func() {
		if len(buf) == 0 {
			return
		}
		batch := buf
		buf = make([]*Log, 0, batchSize)
		if err := LOG_DB.CreateInBatches(batch, len(batch)).Error; err != nil {
			common.SysLog("failed to batch insert logs: " + err.Error())
			// fallback: try one by one
			for _, l := range batch {
				if e := LOG_DB.Create(l).Error; e != nil {
					common.SysLog("failed to insert single log: " + e.Error())
				}
			}
		}
	}

	for {
		select {
		case log, ok := <-logQueue:
			if !ok {
				flush()
				return
			}
			buf = append(buf, log)
			if len(buf) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
