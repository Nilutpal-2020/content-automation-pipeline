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
	"content-automation-pipeline/internal/queue"
	"content-automation-pipeline/internal/scheduler"
	"content-automation-pipeline/internal/store"
	"content-automation-pipeline/pkg/config"
	"content-automation-pipeline/pkg/logger"
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

	if err := store.InitMongoDB(cfg.MongoURI, cfg.MongoDBName); err != nil {
		logger.Log.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	logger.Log.Info("Connected to MongoDB")

	if err := queue.InitRedis(cfg.RedisAddr, cfg.RedisPassword); err != nil {
		logger.Log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Log.Info("Connected to Redis")

	// TODO: Initialize Scheduler/Cron

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Log.Info("Worker is running. Press Ctrl+C to stop.")

	// Wire dependencies
	gen, err := generator.NewGenerator(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to initialize generator", zap.Error(err))
	}
	
	pub := publisher.NewThreadsPublisher(cfg)
	scorer := filter.NewScorer()
	collectors := []collector.Collector{
		collector.NewHackerNewsCollector(),
		collector.NewDevToCollector(),
	}

	pipe := scheduler.NewScheduler(collectors, scorer, gen, pub)

	// Run pipeline on a ticker
	// For example, every 10 minutes in prod, but let's use 1 minute for local testing
	tickerDuration := 1 * time.Minute
	if os.Getenv("ENV") == "production" {
		tickerDuration = 10 * time.Minute
	}
	
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	// Run once immediately on startup
	pipe.RunCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Worker shutting down...")
			return
		case <-ticker.C:
			pipe.RunCycle(ctx)
		}
	}
}
