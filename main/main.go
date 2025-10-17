package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"trade/api"
	"trade/btlLog"
	"trade/config"
	"trade/dao"
	"trade/middleware"
	"trade/routes"
	"trade/routes/RouterSecond"
	"trade/services"
	"trade/services/confs"
	"trade/services/custodyAccount"
	"trade/services/forwardtrans"
	"trade/services/nodemanage"
	"trade/services/servicesrpc"
	"trade/task"
	"trade/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
)

func main() {
	loadConfig, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}
	mode := loadConfig.GinConfig.Mode
	if !(mode == gin.DebugMode || mode == gin.ReleaseMode || mode == gin.TestMode) {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)
	utils.PrintAsciiLogoAndInfo()
	if mode != gin.ReleaseMode {
		utils.PrintTitle(false, "Initialize")
	}

	if err = middleware.InitMysql(); err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return
	}
	if err = middleware.RedisConnect(); err != nil {
		log.Printf("Failed to initialize redis: %v", err)
		return
	}
	if config.GetLoadConfig().IsAutoMigrate {
		if err = dao.Migrate(); err != nil {
			utils.LogError("AutoMigrate error", err)
			return
		}
	}
	utils.PrintTitle(true, "Check Start")
	if !checkStart() {
		return
	}

	go func() {
		_ = confs.ReadConfDefault()
	}()

	services.CheckIfAutoUpdateScheduledTask()
	var jobs []task.Job
	if jobs, err = task.LoadJobs(); err != nil {
		log.Println(err)
		return
	}
	c := cron.New(cron.WithSeconds())
	for _, job := range jobs {

		_, err = c.AddFunc(job.CronExpression, func() {
			task.ExecuteWithLock(job.Name)
		})
		if err != nil {
			log.Printf("Error scheduling job %s: %v\n", job.Name, err)
			continue
		}
	}
	c.Start()
	defer c.Stop()

	if mode != gin.ReleaseMode {
		utils.PrintTitle(true, "Setup Router")
	}
	go middleware.MonitorDatabaseConnections()

	r := routes.SetupRoutes()
	bind := loadConfig.GinConfig.Bind
	port := loadConfig.GinConfig.Port
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    bind + ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	r2 := RouterSecond.SetupRouter()
	r2bind := "127.0.0.1"
	if config.GetConfig().NetWork == "regtest" {
		r2bind = "0.0.0.0"
	}
	localPort := loadConfig.GinConfig.LocalPort
	if localPort == "" {
		localPort = "10080"
	}
	srv2 := &http.Server{
		Addr:    r2bind + ":" + localPort,
		Handler: r2,
	}

	go func() {
		if err := srv2.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			if !checkLitdStatus(1) {
				signalChan <- syscall.SIGTERM
				return
			}
			time.Sleep(time.Second * 10)
		}
	}()

	sig := <-signalChan
	log.Printf("Received signal: %s", sig)

	defer func() {
		closeLitd(loadConfig)
	}()

	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer nodemanage.CloseMineNodes()

	defer func(client *redis.Client) {
		if err = client.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
			return
		} else {
			log.Println("Redis connection closed successfully.")
		}
	}(middleware.Client)

	var db *sql.DB
	if db, err = middleware.DB.DB(); err != nil {
		log.Println(err)
		return
	}
	defer func(db *sql.DB) {
		if err = db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
			return
		} else {
			log.Println("Database connection closed successfully.")
		}
	}(db)

	log.Println("Shutting down the server...")

}

func checkStart() bool {
	cfg := config.GetConfig()

	switch cfg.NetWork {
	case "testnet":
		log.Println("Running on testnet")
	case "mainnet":
		log.Println("Running on mainnet")
	case "regtest":
		log.Println("Running on regtest")
	default:
		log.Println("NetWork need set testnet, mainnet or regtest")
		return false
	}

	if err := btlLog.InitBtlLog(); err != nil {
		log.Printf("Failed to initialize btl log: %v", err)
		return false
	}

	if !checkLitdStatus(1) {

		closeLitd(cfg)
		time.Sleep(time.Second * 5)

		if cfg.ApiConfig.Litd.StartCommand != "" {
			btlLog.START.Info("Start Lit node...")
			cmd := exec.Command("/bin/bash", cfg.ApiConfig.Litd.StartCommand)

			err := cmd.Start()
			if err != nil {
				btlLog.START.Error("Start Lit error: %v", err)
				return false
			}
			if !checkLitdStatus(10) {
				return false
			}
			btlLog.START.Info("Start Lit node success")
		} else {
			return false
		}
	}

	ctx := context.Background()
	if !custodyAccount.CustodyStart(ctx, cfg) {
		return false
	}

	nodemanage.InitNodeManager()
	api.SubscriptionBoxTx()

	err := forwardtrans.LoadMappingCoon()
	if err != nil {
		return false
	}
	forwardtrans.FwdtDaemon()
	return true
}

func checkLitdStatus(retire int) bool {
	for retire > 0 {
		status, err := servicesrpc.LitdStatus()
		if err != nil || status == nil {
			retire--
		}
		if status != nil {
			if status.SubServers["taproot-assets"].GetRunning() {
				return true
			}
			if status.SubServers["taproot-assets"].GetError() != "" {
				btlLog.START.Error("Litd status error: %v", err)
				return false
			}
		}
		if retire != 0 {
			time.Sleep(time.Second * 10)
		}
	}
	btlLog.START.Info("Litd Is not running")
	return false
}

func closeLitd(cfg *config.Config) {

	if cfg.NetWork == "mainnet" && runtime.GOOS != "windows" {
		if _, err := os.Stat("/root/mainnet-trade/not_close_litd"); err == nil {
			return
		}
	}

	if cfg.ApiConfig.Litd.CloseCommand != "" {
		cmd := exec.Command("bash", cfg.ApiConfig.Litd.CloseCommand)

		output, err := cmd.CombinedOutput()
		if err != nil {
			btlLog.START.Error("Close litd error..: %v\n", err.Error())
			return
		}
		btlLog.START.Info("Close litd output: %s\n", output)
		log.Println("Shutting down the Lit node...")
	}
}
