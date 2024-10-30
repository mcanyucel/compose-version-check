package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Files         []FileMapping      `yaml:"files"`
	Notifications NotificationConfig `yaml:"notifications"`
}

type NotificationConfig struct {
	Type         string `yaml:"type"` // "slack", "ntfy", or "debug"
	SlackWebhook string `yaml:"slack_webhook,omitempty"`
	NtfyTopic    string `yaml:"ntfy_topic,omitempty"`
	NtfyServer   string `yaml:"ntfy_server,omitempty"`
	DebugFile    string `yaml:"debug_file,omitempty"` // Path to debug output file
}

// FileMapping represents a local file to source URL mapping
type FileMapping struct {
	LocalPath string `yaml:"local_path"`
	SourceURL string `yaml:"source_url"`
}

// ComposeFile represents the structure we care about in docker-compose
type ComposeFile struct {
	Services map[string]Service `yaml:"services"`
	Path     string             // Local path to the file
	Source   string             // URL source
}

// Service represents a service in docker-compose
type Service struct {
	Image string `yaml:"image"`
}

// Result represents the comparison result for a single file
type Result struct {
	Path           string
	ServiceChanges []ServiceChange
	Error          error
}

// ServiceChange represents a change in service configuration
type ServiceChange struct {
	ServiceName string
	OldImage    string
	NewImage    string
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	debug := flag.Bool("debug", false, "Enable debug mode (print to console)")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override notification type if debug flag is set
	if *debug {
		fmt.Println("Debug mode enabled - notifications will be printed to console")
		config.Notifications.Type = "debug"
	}

	results := checkComposeFilesConcurrently(config.Files)

	message := formatResults(results)
	if err := sendNotification(message, config.Notifications); err != nil {
		fmt.Printf("Error sending notification: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path) // Using os.ReadFile instead of ioutil.ReadFile
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("error parsing config file: %v", err)
	}

	// Validate config
	if len(config.Files) == 0 {
		return Config{}, fmt.Errorf("no file mappings found in config")
	}

	for _, mapping := range config.Files {
		if mapping.LocalPath == "" {
			return Config{}, fmt.Errorf("local_path cannot be empty")
		}
		if mapping.SourceURL == "" {
			return Config{}, fmt.Errorf("source_url cannot be empty")
		}
	}

	// Validate notification config
	if config.Notifications.Type != "slack" && config.Notifications.Type != "ntfy" && config.Notifications.Type != "debug" {
		return Config{}, fmt.Errorf("notification type must be either 'slack', 'ntfy', or 'debug")
	}

	return config, nil
}

func checkComposeFilesConcurrently(mappings []FileMapping) []Result {
	resultChan := make(chan Result, len(mappings))
	var wg sync.WaitGroup

	for _, mapping := range mappings {
		wg.Add(1)
		go func(mapping FileMapping) {
			defer wg.Done()
			result := checkSingleComposeFile(mapping)
			resultChan <- result
		}(mapping)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []Result
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func checkSingleComposeFile(mapping FileMapping) Result {
	result := Result{Path: mapping.LocalPath}

	// Read local file
	localCompose, err := readComposeFile(mapping.LocalPath)
	if err != nil {
		result.Error = fmt.Errorf("error reading local file: %v", err)
		return result
	}

	// Download and parse remote file
	remoteCompose, err := downloadComposeFile(mapping.SourceURL)
	if err != nil {
		result.Error = fmt.Errorf("error downloading remote file: %v", err)
		return result
	}

	// Compare files
	result.ServiceChanges = compareComposeFiles(localCompose, remoteCompose)
	return result
}

func readComposeFile(path string) (ComposeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ComposeFile{}, err
	}

	var compose ComposeFile
	err = yaml.Unmarshal(data, &compose)
	compose.Path = path
	return compose, err
}

func downloadComposeFile(url string) (ComposeFile, error) {
	resp, err := http.Get(url)
	if err != nil {
		return ComposeFile{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body) // Using os.ReadAll instead of ioutil.ReadAll
	if err != nil {
		return ComposeFile{}, err
	}

	var compose ComposeFile
	err = yaml.Unmarshal(data, &compose)
	compose.Source = url
	return compose, err
}

func compareComposeFiles(local, remote ComposeFile) []ServiceChange {
	var changes []ServiceChange

	for serviceName, localService := range local.Services {
		if remoteService, exists := remote.Services[serviceName]; exists {
			if localService.Image != remoteService.Image {
				changes = append(changes, ServiceChange{
					ServiceName: serviceName,
					OldImage:    localService.Image,
					NewImage:    remoteService.Image,
				})
			}
		}
	}

	return changes
}

func formatResults(results []Result) string {
	var buf bytes.Buffer
	buf.WriteString("Docker Compose File Check Results:\n\n")

	for _, result := range results {
		if result.Error != nil {
			buf.WriteString(fmt.Sprintf("❌ Error checking %s: %v\n\n", result.Path, result.Error))
			continue
		}

		if len(result.ServiceChanges) == 0 {
			buf.WriteString(fmt.Sprintf("✅ No changes found for %s\n\n", result.Path))
			continue
		}

		buf.WriteString(fmt.Sprintf("📝 Changes found in %s:\n", result.Path))
		for _, change := range result.ServiceChanges {
			buf.WriteString(fmt.Sprintf("  Service %s:\n", change.ServiceName))
			buf.WriteString(fmt.Sprintf("    Old image: %s\n", change.OldImage))
			buf.WriteString(fmt.Sprintf("    New image: %s\n", change.NewImage))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

func sendNotification(message string, config NotificationConfig) error {
	switch config.Type {
	case "slack":
		return sendSlackNotification(message, config.SlackWebhook)
	case "ntfy":
		return sendNtfyNotification(message, config)
	case "debug":
		return sendDebugNotification(message, config.DebugFile)
	default:
		return fmt.Errorf("unsupported notification type: %s", config.Type)
	}
}

func sendDebugNotification(message string, debugFile string) error {
	// Print to console
	fmt.Printf("\n=== Debug Notification [%s] ===\n", time.Now().Format(time.RFC3339))
	fmt.Println(message)
	fmt.Println("===============================")

	// If debug file is specified, also write to file
	if debugFile != "" {
		// Create timestamp for the filename
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf("%s_%s.log", debugFile, timestamp)

		// Prepare the debug output
		debugOutput := fmt.Sprintf("=== Debug Notification [%s] ===\n%s\n===============================\n",
			time.Now().Format(time.RFC3339),
			message)

		// Write to file
		err := os.WriteFile(filename, []byte(debugOutput), 0644)
		if err != nil {
			return fmt.Errorf("failed to write debug file: %v", err)
		}
		fmt.Printf("Debug output written to: %s\n", filename)
	}

	return nil
}

func sendSlackNotification(message string, webhookURL string) error {
	payload := map[string]string{
		"text": message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

func sendNtfyNotification(message string, config NotificationConfig) error {
	server := config.NtfyServer
	if server == "" {
		server = "https://ntfy.sh"
	}

	url := fmt.Sprintf("%s/%s", server, config.NtfyTopic)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(message))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
