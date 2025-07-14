package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ConfigLogger struct {
	Mode              string `yaml:"mode" mapstructure:"mode"`
	DisableCaller     bool   `yaml:"disable_caller" mapstructure:"disable_caller"`
	DisableStacktrace bool   `yaml:"disable_stacktrace" mapstructure:"disable_stacktrace"`
	Encoding          string `yaml:"encoding" mapstructure:"encoding"`
	Level             string `yaml:"level" mapstructure:"level"`
	ZapType           string `yaml:"zap_type" mapstructure:"zap_type"`

	// Log retention settings
	MaxAge     int    `yaml:"max_age_days" mapstructure:"max_age_days"` // Days to keep logs
	MaxSize    int    `yaml:"max_size_mb" mapstructure:"max_size_mb"`   // Max size in MB
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`   // Max number of backup files
	Compress   bool   `yaml:"compress" mapstructure:"compress"`         // Compress rotated logs
	LogDir     string `yaml:"log_dir" mapstructure:"log_dir"`           // Custom log directory
}

var loggerLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

type ILogger interface {
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
}

type logger struct {
	sugarLogger *zap.SugaredLogger
	logger      *zap.Logger
	key         string
	zapSugar    bool
}

var Logger *logger = &logger{}

func getPath(cfg ConfigLogger) string {
	path := ""
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Use custom log directory if specified, otherwise use default
	if cfg.LogDir != "" {
		path = cfg.LogDir
	} else {
		path = dir + "/logs"
	}

	// Create directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
	}

	// Add path separator
	if strings.Contains(runtime.GOOS, "window") {
		path = path + "\\"
	} else {
		path = path + "/"
	}
	return path
}

func configure(cfg ConfigLogger) zapcore.WriteSyncer {
	path := getPath(cfg)
	timestamp := time.Now().Format("20060102_150405")

	// Set default values if not specified
	if cfg.MaxSize == 0 {
		cfg.MaxSize = 1 // 1 MB default
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 4 // 4 backup files default
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 7 // 7 days default
	}

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path + timestamp + ".log",
		MaxSize:    cfg.MaxSize,    // megabytes
		MaxBackups: cfg.MaxBackups, // number of backup files
		MaxAge:     cfg.MaxAge,     // days
		Compress:   cfg.Compress,   // compress rotated logs
	})
	return zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stderr),
		zapcore.AddSync(w),
	)
}

func GetLogger() *logger {
	return Logger
}

// App Logger constructor
func Newlogger(cfg ConfigLogger) ILogger {
	logLevel, exist := loggerLevelMap[cfg.Level]
	if !exist {
		logLevel = zapcore.DebugLevel
	}

	var encoderCfg zapcore.EncoderConfig
	if cfg.Mode == "pro" {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()

	}
	encoderCfg.LevelKey = "LEVEL"
	encoderCfg.CallerKey = "CALLER"
	encoderCfg.TimeKey = "TIME"
	encoderCfg.NameKey = "NAME"
	encoderCfg.MessageKey = "MESSAGE"
	encoderCfg.EncodeDuration = zapcore.NanosDurationEncoder
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.FunctionKey = "FUNC"
	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, configure(cfg), zap.NewAtomicLevelAt(logLevel))
	loggerzap := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	sugarLogger := loggerzap.Sugar()

	logging := &logger{
		sugarLogger: sugarLogger,
		logger:      loggerzap,
		key:         uuid.NewString(),
		zapSugar:    strings.Contains(cfg.ZapType, "sugar"),
	}

	Logger = logging
	return logging
}

func (l *logger) SetLogID(key string) {
	l.key = key
}

func (l *logger) Debug(args ...interface{}) {
	if l.zapSugar {
		l.sugarLogger.Debug(args...)
		return
	}
	str := fmt.Sprintf("%s", args...)
	fields := []zapcore.Field{
		zap.String("UUID", l.key),
	}
	l.logger.Debug(str, fields...)
}

func (l *logger) Debugf(template string, args ...interface{}) {
	if l.zapSugar {
		str := fmt.Sprintf("UUID:%s, %s", l.key, template)
		l.sugarLogger.Debugf(str, args...)
		return
	}
	str := fmt.Sprintf("%s", args...)
	fields := []zapcore.Field{
		zap.String("UUID", l.key),
	}
	l.logger.Debug(str, fields...)
}

func (l *logger) Info(args ...interface{}) {
	if l.zapSugar {
		l.sugarLogger.Info(args...)
		return
	}
	str := fmt.Sprintf("%s", args...)
	fields := []zapcore.Field{
		zap.String("UUID", l.key),
	}
	l.logger.Info(str, fields...)
}

func (l *logger) Infof(template string, args ...interface{}) {
	if l.zapSugar {
		str := fmt.Sprintf("UUID:%s, %s", l.key, template)
		l.sugarLogger.Infof(str, args...)
		return
	}
	fields := []zapcore.Field{
		zap.String("UUID", l.key),
	}
	l.logger.Info(fmt.Sprintf(template, args...), fields...)
}

func (l *logger) Warn(args ...interface{}) {
	l.sugarLogger.Warn(args...)
}

func (l *logger) Warnf(template string, args ...interface{}) {
	l.sugarLogger.Warnf(template, args...)
}

func (l *logger) Error(args ...interface{}) {
	if l.zapSugar {
		l.sugarLogger.Error(args...)
		return
	}
	str := fmt.Sprintf("%s", args...)
	fields := []zapcore.Field{
		zap.String("UUID", l.key),
	}
	l.logger.Error(str, fields...)
}

func (l *logger) Errorf(template string, args ...interface{}) {
	if l.zapSugar {
		str := fmt.Sprintf("UUID:%s, %s", l.key, template)
		l.sugarLogger.Errorf(str, args...)
		return
	}
	fields := []zapcore.Field{
		zap.String("UUID", l.key),
	}
	l.logger.Error(fmt.Sprintf(template, args...), fields...)
}

func (l *logger) DPanic(args ...interface{}) {
	l.sugarLogger.DPanic(args...)
}

func (l *logger) DPanicf(template string, args ...interface{}) {
	l.sugarLogger.DPanicf(template, args...)
}

func (l *logger) Panic(args ...interface{}) {
	l.sugarLogger.Panic(args...)
}

func (l *logger) Panicf(template string, args ...interface{}) {
	l.sugarLogger.Panicf(template, args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.sugarLogger.Fatal(args...)
}

func (l *logger) Fatalf(template string, args ...interface{}) {
	l.sugarLogger.Fatalf(template, args...)
}
