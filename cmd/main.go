package main

import (
	"context"
	"fmt"
	"gibraltar/config"
	"gibraltar/internal/handlers"
	"gibraltar/internal/services"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	tester := services.NewVlessTestService(config.TestURL)
	cache := services.NewCache()
	dataSource := services.NewUrlParser(config.VlessSecureConfigsURLs, config.CIDRWhitelistURL, config.URLsWhitelistURL)
	CIDRlist, err := dataSource.ParseSubnets()
	if err != nil {
		panic(fmt.Errorf("Can't get CIDR whitelist: %s", err))
	}
	allowedSNIs, err := dataSource.ParseSNIs()
	if err != nil {
		panic(fmt.Errorf("Can't get SNI whitelist: %s", err))
	}
	filter := services.NewConfigFilter(CIDRlist, allowedSNIs)
	updater := services.NewConfigUpdater(cache, filter, tester, dataSource)
	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		updateTick := time.NewTicker(config.UpdateCooldown)
		defer updateTick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-updateTick.C:
				tick.Reset(time.Second)
				log.Println("Configs updating...")
				if err := updater.AddConfigsToCacheFromSource(); err != nil {
					log.Println(err)
				}
				updateTick.Reset(config.UpdateCooldown)
			case <-tick.C:
				tick.Reset(time.Minute)
				if err := updater.AddAvailableConfigsToCache(); err != nil {
					log.Println(err)
				}
			}
		}
	}()

	cfgHandler := handlers.NewConfigsHandler(cache)
	router := gin.Default()
	router.GET("/configs", cfgHandler.CurrentAvailableConfigs)
	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router.Handler(),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("http server error:", err)
		}

	}()

	pprofSrv := &http.Server{
		Addr: "127.0.0.1:6060",
	}

	go func() {
		if err := pprofSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("pprof server error:", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Println("Termination signal received, shutting down...")
	cancel()
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	_ = srv.Shutdown(ctxShutdown)
	_ = pprofSrv.Shutdown(ctxShutdown)
	wg.Wait()

}
