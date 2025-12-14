package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
}

// Weighted levels - higher weight = more frequent
var levelWeights = []struct {
	level  string
	weight int
}{
	{"DEBUG", 30},
	{"INFO", 40},
	{"WARN", 15},
	{"ERROR", 10},
	{"CRITICAL", 5},
}

var sampleMessages = map[string][]string{
	"DEBUG": {
		"Processing request payload",
		"Cache hit for key: user_session_%d",
		"Database query executed in %dms",
		"Validating input parameters",
		"Loading configuration from environment",
		"Memory allocation: %d bytes",
		"Goroutine count: %d",
		"Request headers parsed",
		"Response serialization complete",
		"Middleware chain executed",
	},
	"INFO": {
		"User login successful: user_%d",
		"Order #%d created successfully",
		"Payment processed for $%d.99",
		"Email sent to user%d@example.com",
		"New user registration completed",
		"API request completed in %dms",
		"Cache refreshed for %d items",
		"Scheduled task completed",
		"File uploaded: %d bytes",
		"Session created for user_%d",
	},
	"WARN": {
		"High memory usage detected: %d%%",
		"Slow database query: %dms",
		"Rate limit approaching: %d/100 requests",
		"Deprecated API endpoint called",
		"Retry attempt %d/3 for external service",
		"Connection pool running low: %d available",
		"Disk usage at %d%%",
		"Response time degraded: %dms",
		"Certificate expires in %d days",
		"Queue backlog: %d items",
	},
	"ERROR": {
		"Failed to connect to database: timeout after %ds",
		"Payment gateway timeout for order #%d",
		"Invalid authentication token",
		"File upload failed: %d bytes rejected",
		"External API returned 500 error",
		"Request validation failed: %d errors",
		"Database transaction rollback",
		"Service unavailable: retry in %ds",
		"Connection refused to port %d",
		"Null pointer exception in handler",
	},
	"CRITICAL": {
		"Database connection pool exhausted",
		"Out of memory - service degraded",
		"Security breach detected from IP %d.%d.%d.%d",
		"Data corruption in primary storage",
		"System health check failed",
		"Cluster node %d unresponsive",
		"Disk failure imminent: %d bad sectors",
		"Master database failover initiated",
		"Service crash: exit code %d",
		"Replication lag: %d seconds",
	},
}

var sources = []string{
	"api-server",
	"worker",
	"scheduler",
	"auth-service",
	"payment-service",
	"notification-service",
	"user-service",
	"order-service",
	"gateway",
	"cache-service",
}

func main() {
	apiURL := flag.String("url", "http://localhost:3000/api/v1/logs", "API endpoint URL")
	apiKey := flag.String("key", "", "API key (required)")
	duration := flag.Duration("duration", 0, "Duration to run (e.g., 1h, 30m, 1h30m)")
	minInterval := flag.Duration("min", 5*time.Second, "Minimum interval between logs")
	maxInterval := flag.Duration("max", 30*time.Second, "Maximum interval between logs")
	quiet := flag.Bool("q", false, "Quiet mode - minimal output")

	flag.Parse()

	if *apiKey == "" {
		fmt.Println("Log Generator - Generate realistic random logs")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Run for 1 hour with random intervals (5-30s)")
		fmt.Println("  loggen -key YOUR_API_KEY -duration 1h")
		fmt.Println("")
		fmt.Println("  # Run for 30 minutes with faster logs (2-10s)")
		fmt.Println("  loggen -key YOUR_API_KEY -duration 30m -min 2s -max 10s")
		fmt.Println("")
		fmt.Println("  # Run indefinitely (Ctrl+C to stop)")
		fmt.Println("  loggen -key YOUR_API_KEY")
		fmt.Println("")
		fmt.Println("  # Quiet mode")
		fmt.Println("  loggen -key YOUR_API_KEY -duration 1h -q")
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var endTime time.Time
	if *duration > 0 {
		endTime = time.Now().Add(*duration)
		if !*quiet {
			fmt.Printf("Generating logs for %v (until %s)\n", *duration, endTime.Format("15:04:05"))
			fmt.Printf("Interval: %v - %v (random)\n", *minInterval, *maxInterval)
			fmt.Println("Press Ctrl+C to stop early.")
		}
	} else {
		if !*quiet {
			fmt.Println("Generating logs indefinitely. Press Ctrl+C to stop.")
		}
	}

	logCount := 0
	startTime := time.Now()

	for {
		select {
		case <-sigChan:
			fmt.Printf("\n\nStopped. Generated %d logs in %v\n", logCount, time.Since(startTime).Round(time.Second))
			return
		default:
			// Check if duration exceeded
			if *duration > 0 && time.Now().After(endTime) {
				fmt.Printf("\n\nDuration completed. Generated %d logs in %v\n", logCount, *duration)
				return
			}

			// Send log
			sendLog(*apiURL, *apiKey, *quiet)
			logCount++

			// Random interval between min and max
			interval := *minInterval + time.Duration(rand.Int63n(int64(*maxInterval-*minInterval)))
			time.Sleep(interval)
		}
	}
}

func sendLog(apiURL, apiKey string, quiet bool) {
	// Pick weighted random level
	level := pickWeightedLevel()

	// Pick random message with dynamic values
	message := generateMessage(level)

	// Pick random source
	source := sources[rand.Intn(len(sources))]

	log := LogEntry{
		Level:     level,
		Message:   message,
		Source:    source,
		Timestamp: time.Now().Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"generator":  "loggen",
			"hostname":   getHostname(),
			"request_id": fmt.Sprintf("req_%d", rand.Intn(1000000)),
			"pid":        rand.Intn(65535),
		},
	}

	body, err := json.Marshal(log)
	if err != nil {
		if !quiet {
			fmt.Printf("Error marshaling log: %v\n", err)
		}
		return
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		if !quiet {
			fmt.Printf("Error creating request: %v\n", err)
		}
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if !quiet {
			fmt.Printf("Error sending request: %v\n", err)
		}
		return
	}
	defer resp.Body.Close()

	if !quiet {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		levelColor := getLevelColor(level)
		timestamp := time.Now().Format("15:04:05")

		if resp.StatusCode == 201 {
			fmt.Printf("[%s] %s%-8s%s %s\n", timestamp, levelColor, level, "\033[0m", message)
		} else {
			fmt.Printf("[%s] %sFAILED%s  %v\n", timestamp, "\033[31m", "\033[0m", result["error"])
		}
	}
}

func pickWeightedLevel() string {
	totalWeight := 0
	for _, lw := range levelWeights {
		totalWeight += lw.weight
	}

	r := rand.Intn(totalWeight)
	for _, lw := range levelWeights {
		r -= lw.weight
		if r < 0 {
			return lw.level
		}
	}
	return "INFO"
}

func generateMessage(level string) string {
	messages := sampleMessages[level]
	template := messages[rand.Intn(len(messages))]

	// Count %d placeholders and replace with random values
	count := strings.Count(template, "%d")
	if count == 0 {
		return template
	}

	args := make([]interface{}, count)
	for i := 0; i < count; i++ {
		args[i] = rand.Intn(9999) + 1
	}

	return fmt.Sprintf(template, args...)
}

func getLevelColor(level string) string {
	switch level {
	case "DEBUG":
		return "\033[36m" // Cyan
	case "INFO":
		return "\033[32m" // Green
	case "WARN":
		return "\033[33m" // Yellow
	case "ERROR":
		return "\033[31m" // Red
	case "CRITICAL":
		return "\033[35m" // Magenta
	default:
		return "\033[0m"
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
