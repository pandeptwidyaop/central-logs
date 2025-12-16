package worker

import (
	"central-logs/internal/models"
	"central-logs/internal/queue"
	"context"
	"log"
	"time"
)

// NotificationConsumer processes notification jobs from Redis queue
type NotificationConsumer struct {
	redisClient *queue.RedisClient
	notifier    *Notifier
	channelRepo *models.ChannelRepository
	logRepo     *models.LogRepository
	stopChan    chan struct{}
}

// NewNotificationConsumer creates a new notification consumer
func NewNotificationConsumer(
	redisClient *queue.RedisClient,
	notifier *Notifier,
	channelRepo *models.ChannelRepository,
	logRepo *models.LogRepository,
) *NotificationConsumer {
	return &NotificationConsumer{
		redisClient: redisClient,
		notifier:    notifier,
		channelRepo: channelRepo,
		logRepo:     logRepo,
		stopChan:    make(chan struct{}),
	}
}

// Start begins consuming notification jobs from the queue
func (nc *NotificationConsumer) Start(workers int) {
	log.Printf("Starting %d notification workers...", workers)

	for i := 0; i < workers; i++ {
		go nc.worker(i)
	}
}

// Stop signals all workers to stop
func (nc *NotificationConsumer) Stop() {
	log.Println("Stopping notification workers...")
	close(nc.stopChan)
}

// worker is the main loop for processing notification jobs
func (nc *NotificationConsumer) worker(id int) {
	log.Printf("Notification worker #%d started", id)

	ctx := context.Background()

	for {
		select {
		case <-nc.stopChan:
			log.Printf("Notification worker #%d stopped", id)
			return
		default:
			// Dequeue with timeout to allow graceful shutdown
			job, err := nc.redisClient.DequeueNotification(ctx, 5*time.Second)
			if err != nil {
				log.Printf("Worker #%d: Error dequeuing notification: %v", id, err)
				time.Sleep(time.Second)
				continue
			}

			if job == nil {
				// No job available, continue
				continue
			}

			// Process the notification job
			nc.processJob(job)
		}
	}
}

// processJob handles a single notification job
func (nc *NotificationConsumer) processJob(job *queue.NotificationJob) {
	// Get the channel
	channel, err := nc.channelRepo.GetByID(job.ChannelID)
	if err != nil {
		log.Printf("Failed to get channel %s: %v", job.ChannelID, err)
		return
	}

	if channel == nil {
		log.Printf("Channel %s not found", job.ChannelID)
		return
	}

	if !channel.IsActive {
		log.Printf("Channel %s is inactive, skipping", job.ChannelID)
		return
	}

	// Get the log entry
	logEntry, err := nc.logRepo.GetByID(job.LogID)
	if err != nil {
		log.Printf("Failed to get log %s: %v", job.LogID, err)
		return
	}

	if logEntry == nil {
		log.Printf("Log %s not found", job.LogID)
		return
	}

	// Send notification based on channel type
	switch channel.Type {
	case models.ChannelTypeTelegram:
		nc.notifier.sendTelegram(channel, logEntry)
	case models.ChannelTypeDiscord:
		nc.notifier.sendDiscord(channel, logEntry)
	case models.ChannelTypePush:
		nc.notifier.sendPush(channel, logEntry)
	default:
		log.Printf("Unknown channel type: %s", channel.Type)
	}
}
