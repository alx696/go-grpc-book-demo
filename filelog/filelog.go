package filelog

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"
)

var e error
var logFile *os.File
var l *log.Logger
var fileTag string

func changeLogFile() {
	// 关闭之前文件
	if logFile != nil {
		_ = logFile.Close()
	}

	// 创建日志文件
	logFile, e = os.CreateTemp("", fmt.Sprintf(`%s-%s-`, fileTag, time.Now().Format("20060102_150405")))
	if e != nil {
		log.Fatalln(e)
	}

	// 更新日志器
	if l == nil {
		l = log.New(io.MultiWriter(os.Stdout, logFile), "", log.Ldate|log.Ltime|log.Lmicroseconds)
	} else {
		l.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}
}

func print(tag string, v ...any) {
	funcName := "?"
	pc, filename, line, ok := runtime.Caller(2)
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	} else {
		funcName = filename
	}

	l.Println(append([]any{tag, fmt.Sprint(funcName, ":", line)}, v...)...)
}

// Init 初始, 在 main.go 中调用一次
//
// fileTagArg 文件标签:可用于区分日志文件
func Init(fileTagArg string) {
	// 文件标记
	fileTag = fileTagArg

	// 更换日志文件
	changeLogFile()

	// 安排定时任务
	cs := cron.New(cron.WithSeconds())
	cs.AddFunc("0 0 * * * *", changeLogFile)
	cs.Start()
}

// Over 结束, 在 main.go 中调用一次
func Over() {
	// 关闭之前文件
	if logFile != nil {
		_ = logFile.Close()
	}
}

// Error 错误
func Error(v ...any) {
	print("错误", v)
}

// Warn 警告
func Warn(v ...any) {
	print("警告", v)
}

// Info 信息
func Info(v ...any) {
	print("信息", v)
}

// Debug 调试
func Debug(v ...any) {
	print("调试", v)
}
