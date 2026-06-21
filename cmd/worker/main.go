package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"content-automation-pipeline/internal/collector"
	"content-automation-pipeline/internal/filter"
	"content-automation-pipeline/internal/generator"
	"content-automation-pipeline/internal/publisher"
	"content-automation-pipeline/internal/scheduler"
	"content-automation-pipeline/pkg/config"
	"content-automation-pipeline/pkg/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	if err := logger.InitLogger(cfg.LogLevel); err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Content Automation Pipeline", zap.String("env", os.Getenv("ENV")))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Log.Info("Worker is running. Press Ctrl+C to stop.")

	// Wire dependencies
	gen, err := generator.NewGenerator(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to initialize generator", zap.Error(err))
	}
	
	pub := publisher.NewNotionPublisher(cfg)
	scorer := filter.NewScorer()
	collectors := []collector.Collector{
		collector.NewHackerNewsCollector(),
		collector.NewDevToCollector(),
	}

	pipe := scheduler.NewScheduler(collectors, scorer, gen, pub)

	// Set up cron scheduler
	c := cron.New(cron.WithLocation(time.Local))

	// Run every day at 8:00 AM in production, or every minute locally
	cronSpec := "0 8 * * *"
	if os.Getenv("ENV") != "production" {
		cronSpec = "* * * * *"
	}

	_, err = c.AddFunc(cronSpec, func() {
		cycleCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		pipe.RunCycle(cycleCtx)
	})
	if err != nil {
		logger.Log.Fatal("Failed to setup cron schedule", zap.Error(err))
	}

	c.Start()

	<-ctx.Done()
	logger.Log.Info("Worker shutting down...")
	
	cronCtx := c.Stop()
	<-cronCtx.Done()
}
