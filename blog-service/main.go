package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-programming-tour-book/blog-service/pkg/tracer"

	"github.com/gin-gonic/gin"
	"github.com/go-programming-tour-book/blog-service/global"
	"github.com/go-programming-tour-book/blog-service/internal/model"
	"github.com/go-programming-tour-book/blog-service/internal/routers"
	"github.com/go-programming-tour-book/blog-service/pkg/logger"
	"github.com/go-programming-tour-book/blog-service/pkg/setting"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	port      string
	runMode   string
	config    string
	isVersion bool

	buildTime    string
	buildVersion string
	gitCommitID  string
)

func init() {
	err := setupSetting()
	if err != nil {
		log.Fatalf("init.setupSetting err: %v", err)
	}
	err = setupLogger()
	if err != nil {
		log.Fatalf("init.setupLogger err: %v", err)
	}
	setupFlag()
}

func setupFlag() error {
	flag.StringVar(&port, "port", "", "启动端口")
	flag.StringVar(&runMode, "mode", "", "启动模式")
	flag.StringVar(&config, "config", "configs/", "指定要使用的配置文件路径")
	flag.BoolVar(&isVersion, "version", false, "编译信息")
	flag.Parse()

	return nil
}

// @title 博客系统
// @version 1.0
// @description Go编程之旅
// @termsOfService https://github.com/go-programming-tour-book
func main() {
	//ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	//global.Logger.Infof(ctx, "%s: go-programming-tour-book %s", "eddycjy","blog-service")

	if isVersion {
		fmt.Printf("build_time: %s\n", buildTime)
		fmt.Printf("build_version: %s\n", buildVersion)
		fmt.Printf("git_commit_id: %s\n", gitCommitID)
		return
	}

	gin.SetMode(global.ServerSetting.RunMode)
	router := routers.NewRouter()
	s := &http.Server{
		Addr:           ":" + global.ServerSetting.HttpPort,
		Handler:        router,
		ReadTimeout:    global.ServerSetting.ReadTimeout,
		WriteTimeout:   global.ServerSetting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("s.ListenAndServe err: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal)
	// 接收syscall.SIGINT和syscall.SIGTERM信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shuting down server...")

	// 最大时间控制，用于通知该服务端它有5秒的时间来处理原有的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}

func setupSetting() error {
	setting, err := setting.NewSetting(strings.Split(config, ",")...)
	if err != nil {
		return err
	}
	err = setting.ReadSection("Server", &global.ServerSetting)
	if err != nil {
		return err
	}
	err = setting.ReadSection("App", &global.AppSetting)
	if err != nil {
		return err
	}
	err = setting.ReadSection("Database", &global.DatabaseSetting)
	if err != nil {
		return err
	}
	err = setting.ReadSection("JWT", &global.JWTSetting)
	if err != nil {
		return err
	}
	err = setting.ReadSection("Email", &global.EmailSetting)
	if err != nil {
		return err
	}
	err = setupTracer()
	if err != nil {
		log.Fatalf("init.setupTracer err:%v", err)
	}

	global.ServerSetting.ReadTimeout *= time.Second
	global.ServerSetting.WriteTimeout *= time.Second
	global.JWTSetting.Expire *= time.Second

	return nil
}

func setupLogger() error {
	global.Logger = logger.NewLogger(&lumberjack.Logger{
		Filename:  global.AppSetting.LogSavePath + "/" + global.AppSetting.LogFileName + global.AppSetting.LogFileExt,
		MaxSize:   600,
		MaxAge:    10,
		LocalTime: true,
	}, "", log.LstdFlags).WithCaller(2)
	return nil
}

func setupDBEngine() error {
	var err error
	global.DBEngine, err = model.NewDBEngine(global.DatabaseSetting)
	if err != nil {
		return err
	}

	return nil
}

func setupTracer() error {
	jaegerTracer, _, err := tracer.NewJaegerTracer(
		"blog-service",
		"127.0.0.1:6831",
	)
	if err != nil {
		return err
	}
	global.Tracer = jaegerTracer
	return nil
}
