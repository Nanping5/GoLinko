package zlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// HourlyRotateWriter 实现按小时轮转的日志写入器
type HourlyRotateWriter struct {
	logDir      string
	currentFile *os.File
	currentHour string
	mu          sync.Mutex
}

// NewHourlyRotateWriter 创建一个新的按小时轮转的写入器
func NewHourlyRotateWriter(logDir string) (*HourlyRotateWriter, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	writer := &HourlyRotateWriter{
		logDir: logDir,
	}

	// 初始化当前小时的日志文件
	if err := writer.rotate(); err != nil {
		return nil, err
	}

	return writer, nil
}

// Write 实现 io.Writer 接口
func (w *HourlyRotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 检查是否需要轮转
	currentHour := time.Now().Format("2006-01-02-15")
	if currentHour != w.currentHour {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	return w.currentFile.Write(p)
}

// rotate 轮转到新的日志文件
func (w *HourlyRotateWriter) rotate() error {
	// 关闭旧文件
	if w.currentFile != nil {
		if err := w.currentFile.Close(); err != nil {
			return fmt.Errorf("关闭旧日志文件失败: %w", err)
		}
	}

	// 生成新的文件名
	w.currentHour = time.Now().Format("2006-01-02-15")
	filename := fmt.Sprintf("log-%s.log", w.currentHour)
	logPath := filepath.Join(w.logDir, filename)

	// 打开新文件
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	w.currentFile = file
	return nil
}

// Close 关闭当前日志文件
func (w *HourlyRotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.currentFile != nil {
		return w.currentFile.Close()
	}
	return nil
}

func init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	//设置日志时间格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//日志encoder还是JsonEncoder，只不过把日志格式化为JSON
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建按小时轮转的写入器
	logDir := "logs"
	rotateWriter, err := NewHourlyRotateWriter(logDir)
	if err != nil {
		panic(fmt.Sprintf("初始化日志系统失败: %v", err))
	}

	//Zap 需要一个 WriteSyncer 接口
	fileWriteSyncer := zapcore.AddSync(rotateWriter)

	//zapcore.NewTee 可以将多个 Core 组合成一个 Core，这样日志就会同时输出到多个地方
	core := zapcore.NewTee(
		//输出到控制台，日志级别为 Debug 及以上
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		//输出到文件，日志级别为 Debug 及以上
		zapcore.NewCore(encoder, fileWriteSyncer, zapcore.DebugLevel),
	)
	//启用调用者信息，日志中会包含调用者的文件名和行号
	logger = zap.New(core, zap.AddCaller())
}

//废案
// // 获取调用者信息，包含文件名和行号
// func GetCallerInfoForLog() (callerFields []zap.Field) {
// 	pc, file, line, ok := runtime.Caller(2) //回溯两层，获取调用者信息
// 	if !ok {
// 		return []zap.Field{}
// 	}
// 	funcName := runtime.FuncForPC(pc).Name() //获取函数名
// 	funcName = path.Base(funcName) //只保留函数名最后一部分
// 	callerFields = append(callerFields, zap.String("func", funcName), zap.String("file", file), zap.Int("line", line))
// 	return callerFields
// }

// 提供一个全局的日志记录器，供其他包调用
func GetLogger() *zap.Logger {
	return logger
}
