package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	facker "kafka-consumer/application/faker"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azure/go-ntlmssp"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"kafka-consumer/application/constant"
	"kafka-consumer/application/logger"
	"kafka-consumer/application/model"
	"kafka-consumer/config"
)

// BartenderPrinterJob represents a print job
type BartenderPrinterJob struct {
	Filename            string
	DocumentFilePath    string
	ConnectionSetupPath string
	RetryCount          int
	MaxRetries          int
	CreatedAt           time.Time
	LastAttemptAt       time.Time
}

// KafkaService contains all dependencies needed for Kafka operations
type KafkaService struct {
	logger logger.ILogger
	config *config.Config

	// Bartender Printer optimization
	bartenderClient   *http.Client
	rateLimiter       *rate.Limiter
	jobQueue          chan *BartenderPrinterJob
	workerCount       int
	maxConcurrentJobs int
	healthCheckTicker *time.Ticker
	bartenderHealthy  bool
	healthMutex       sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	// Semaphore to ensure only 1 API call at a time
	semaphore chan struct{}
}

// NewKafkaService creates a new KafkaService instance
func NewKafkaService(logger logger.ILogger, config *config.Config) *KafkaService {
	ctx, cancel := context.WithCancel(context.Background())

	// Create optimized HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Wrap with NTLM authentication
	ntlmTransport := ntlmssp.Negotiator{
		RoundTripper: transport,
	}

	client := &http.Client{
		Transport: ntlmTransport,
		Timeout:   30 * time.Second, // Add timeout
	}

	// Configure rate limiting from config or use defaults
	rateLimit := config.BartenderPrinterAPI.RateLimit
	if rateLimit <= 0 {
		rateLimit = 10 // Default 10 RPS
	}
	rateLimiter := rate.NewLimiter(rate.Limit(rateLimit), rateLimit/2) // Burst = rate/2

	// Configure job queue and workers from config or use defaults
	workerCount := config.BartenderPrinterAPI.WorkerCount
	//if workerCount <= 0 {
	//	workerCount = 3 // Default 3 workers
	//}

	queueSize := config.BartenderPrinterAPI.QueueSize
	if queueSize <= 0 {
		queueSize = 100 // Default queue size
	}

	ks := &KafkaService{
		logger:            logger,
		config:            config,
		bartenderClient:   client,
		rateLimiter:       rateLimiter,
		jobQueue:          make(chan *BartenderPrinterJob, queueSize),
		workerCount:       workerCount,
		maxConcurrentJobs: workerCount, // Use worker count as max concurrent
		bartenderHealthy:  true,
		ctx:               ctx,
		cancel:            cancel,
		semaphore:         make(chan struct{}, 1), // Only 1 API call at a time
	}

	//// Start health check goroutine
	//ks.startHealthCheck()

	// Start worker goroutines
	ks.startWorkers()

	return ks
}

// startHealthCheck starts periodic health check of Bartender Printer
func (ks *KafkaService) startHealthCheck() {
	ks.healthCheckTicker = time.NewTicker(30 * time.Second) // Check every 30 seconds

	ks.wg.Add(1)
	go func() {
		defer ks.wg.Done()
		for {
			select {
			case <-ks.ctx.Done():
				return
			case <-ks.healthCheckTicker.C:
				ks.checkBartenderHealth()
			}
		}
	}()
}

// checkBartenderHealth checks if Bartender Printer is healthy
func (ks *KafkaService) checkBartenderHealth() {
	// Simple health check - try to connect to Bartender API
	req, err := http.NewRequest("GET", ks.config.BartenderPrinterAPI.URL, nil)
	if err != nil {
		ks.setHealthStatus(false)
		return
	}

	req.Header.Set("accept", "application/json")
	req.SetBasicAuth(ks.config.BartenderPrinterAPI.Username, ks.config.BartenderPrinterAPI.Password)

	ctx, cancel := context.WithTimeout(ks.ctx, 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := ks.bartenderClient.Do(req)
	if err != nil {
		ks.logger.Warnf("Bartender health check failed: %v", err)
		ks.setHealthStatus(false)
		return
	}
	defer resp.Body.Close()

	// Consider healthy if we get any response (even error responses)
	ks.setHealthStatus(resp.StatusCode < 500)
}

// setHealthStatus updates the health status with thread safety
func (ks *KafkaService) setHealthStatus(healthy bool) {
	ks.healthMutex.Lock()
	defer ks.healthMutex.Unlock()
	ks.bartenderHealthy = healthy
}

// isHealthy checks if Bartender is healthy
func (ks *KafkaService) isHealthy() bool {
	ks.healthMutex.RLock()
	defer ks.healthMutex.RUnlock()
	return ks.bartenderHealthy
}

// startWorkers starts worker goroutines to process print jobs
func (ks *KafkaService) startWorkers() {
	if ks.config.BartenderPrinterAPI.SequentialMode {
		// Sequential mode: only 1 worker to ensure 1 API call at a time
		ks.logger.Info("Starting in SEQUENTIAL mode - only 1 API call at a time")
		ks.wg.Add(1)
		go func() {
			defer ks.wg.Done()
			ks.worker(0)
		}()
	} else {
		// Parallel mode: multiple workers for concurrent API calls
		ks.logger.Infof("Starting in PARALLEL mode with %d workers", ks.workerCount)
		for i := 0; i < ks.workerCount; i++ {
			ks.wg.Add(1)
			go func(workerID int) {
				defer ks.wg.Done()
				ks.worker(workerID)
			}(i)
		}
	}
}

// worker processes print jobs from the queue
func (ks *KafkaService) worker(workerID int) {
	ks.logger.Infof("Bartender worker %d started", workerID)

	for {
		select {
		case <-ks.ctx.Done():
			ks.logger.Infof("Bartender worker %d shutting down", workerID)
			return
		case job := <-ks.jobQueue:
			ks.processPrintJob(job, workerID)
		}
	}
}

// processPrintJob processes a single print job with retry logic
func (ks *KafkaService) processPrintJob(job *BartenderPrinterJob, workerID int) {
	ks.semaphore <- struct{}{} // Acquire semaphore at the start
	defer func() { <-ks.semaphore }()

	for {
		job.LastAttemptAt = time.Now()
		// Wait for rate limiter
		if err := ks.rateLimiter.Wait(ks.ctx); err != nil {
			ks.logger.Errorf("Rate limiter error: %v", err)
			return
		}

		// Process the job
		resp, err := ks.callBartenderPrinterAPI(job.Filename, false, job.DocumentFilePath, job.ConnectionSetupPath)
		if err != nil {
			ks.logger.Errorf("Worker %d failed to process job %s: %v", workerID, job.Filename, err)
			// Check if we can still retry
			if job.RetryCount < job.MaxRetries {
				job.RetryCount++
				backoff := time.Duration(job.RetryCount) * 2 * time.Second
				ks.logger.Infof("Retrying job %s in %v (attempt %d/%d)", job.Filename, backoff, job.RetryCount, job.MaxRetries)

				// Schedule retry with backoff
				go func() {
					time.Sleep(backoff)
					select {
					case ks.jobQueue <- job:
						ks.logger.Infof("Requeued job %s for retry %d", job.Filename, job.RetryCount)
					default:
						ks.logger.Errorf("Job queue full, dropping retry job: %s", job.Filename)
					}
				}()
			} else {
				// Max retries reached, log and stop
				ks.logger.Errorf("Job %s FAILED PERMANENTLY after %d retries - STOPPING RETRY", job.Filename, job.MaxRetries)
				time.Sleep(5 * time.Second)
				// Job is now considered failed permanently, no more retries
				return
			}
		} else {
			ks.logger.Infof("Worker %d successfully processed job: %s", workerID, job.Filename)
			if resp != nil && resp.Status == "WaitingToRun" {
				time.Sleep(15 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
			return
		}
	}
}

// Close gracefully shuts down the KafkaService
func (ks *KafkaService) Close() error {
	ks.logger.Info("Shutting down KafkaService...")

	// Stop health check
	if ks.healthCheckTicker != nil {
		ks.healthCheckTicker.Stop()
	}

	// Cancel context to stop all goroutines
	ks.cancel()

	// Wait for all goroutines to finish
	ks.wg.Wait()

	// Close job queue
	close(ks.jobQueue)

	ks.logger.Info("KafkaService shutdown complete")
	return nil
}

// StartConsumer starts the Kafka consumer
func (ks *KafkaService) StartConsumer() error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": ks.config.Kafka.BootstrapServers,
		"group.id":          ks.config.Kafka.GroupID,
		"auto.offset.reset": ks.config.Kafka.AutoOffsetReset,
	})
	if err != nil {
		return err
	}
	defer func(c *kafka.Consumer) {
		err := c.Close()
		if err != nil {
			ks.logger.Errorf("Failed to close consumer: %s", err)
		}
	}(c)

	err = c.SubscribeTopics([]string{ks.config.ConsumerTopicInfo.TopicBomBartenderPrinter}, nil)
	if err != nil {
		ks.logger.Errorf("Failed to subscribe to topic: %s", err)
		return err
	}

	ks.logger.Infof("[Consumer Application Started] Listening to server: %s, topic: %s",
		ks.config.Kafka.BootstrapServers, ks.config.ConsumerTopicInfo.TopicBomBartenderPrinter)

	// Start consuming messages
	for {
		msg, err := c.ReadMessage(-1)
		if err != nil {
			ks.logger.Errorf("Consumer error: %v", err)
			continue
		}

		if err := ks.processMessage(msg); err != nil {
			ks.logger.Errorf("Error processing message: %v", err)
			continue
		}
	}
}

// processMessage handles individual Kafka messages
func (ks *KafkaService) processMessage(msg *kafka.Message) error {
	var productPrinterMsg model.ProductPrinterMsgKafkaRequest
	if err := json.Unmarshal(msg.Value, &productPrinterMsg); err != nil {
		ks.logger.Errorf("JSON unmarshal error: %v", err)
		return err
	}

	if len(productPrinterMsg.Products) == 0 {
		ks.logger.Warn("No products to export")
		return nil
	}

	documentFilePath, err := ks.getDocumentFilePath(productPrinterMsg.Template)
	if err != nil {
		ks.logger.Errorf("Error getting document file path: %v", err)
		return err
	}

	connectionSetupFile, err := ks.getConnectionFilePath(productPrinterMsg.Template)
	if err != nil {
		ks.logger.Errorf("Error getting connection setup file path: %v", err)
		return err
	}

	now := time.Now()
	quantity := len(productPrinterMsg.Products)
	fileFormat := constant.FileTypeTxt
	filename := "test" + "_" + now.Format("20060102_150405") + "_" + strconv.Itoa(quantity) + "." + string(fileFormat)
	filepath := ks.config.FileSharePath + string(os.PathSeparator) + filename

	ks.populateAndRemakeProducts(productPrinterMsg.Products)

	if err := ks.exportProducts(productPrinterMsg.Products, filepath, fileFormat); err != nil {
		ks.logger.Errorf("Export error: %v", err)
		return err
	}
	ks.logger.Infof("Exported products to %s at time: %s", filepath, now)

	job := &BartenderPrinterJob{
		Filename:            filename,
		DocumentFilePath:    documentFilePath,
		ConnectionSetupPath: connectionSetupFile,
		RetryCount:          0,
		MaxRetries:          ks.config.BartenderPrinterAPI.MaxRetries,
		CreatedAt:           time.Now(),
		LastAttemptAt:       time.Now(),
	}

	select {
	case ks.jobQueue <- job:
		ks.logger.Infof("Enqueued print job for %s", filename)
	default:
		ks.logger.Errorf("Job queue full, dropping print job for %s", filename)
	}

	return nil
}

// exportProducts exports products to file based on format
func (ks *KafkaService) exportProducts(products []*model.Product, filename string, filetype constant.FileType) error {
	switch filetype {
	case constant.FileTypeTxt:
		return ks.exportProductDataToTxtFile(products, filename)
	case constant.FileTypeCsv:
		return ks.exportProductDataToCsvFile(products, filename)
	default:
		return fmt.Errorf("unsupported file type: %s", filetype)
	}
}

// exportProductDataToTxtFile exports product data to TXT file
func (ks *KafkaService) exportProductDataToTxtFile(products []*model.Product, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			ks.logger.Errorf("Failed to close file: %s", err)
		}
	}(file)

	// Write header
	headerFields := []string{"name", "code", "color", "material", "manufacture_office", "manufacture_company", "manufacture_date", "us_size", "vn_size", "uk_size", "gender", "attribute", "size_available", "qr_code", "rfid_barcode", "price", "currency"}
	headerLine := strings.Join(headerFields, ";") + "\n"
	_, err = file.WriteString(headerLine)
	if err != nil {
		return err
	}

	// Write data
	for _, p := range products {
		line := fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s\n",
			p.Name, p.Code, p.Color, p.Material, p.ManufactureOffice, p.ManuFactureCompany, p.ManufactureDate, p.USSize,
			p.VNSize, p.UKSize, p.Gender, p.Attribute, p.SizeAvailable,
			p.QrCode, p.RfidBarcode, p.Price, p.Currency)
		_, err = file.WriteString(line)
		if err != nil {
			return err
		}
	}
	return nil
}

// exportProductDataToCsvFile exports product data to CSV file
func (ks *KafkaService) exportProductDataToCsvFile(products []*model.Product, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			ks.logger.Errorf("Failed to close file: %s", err)
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Define header fields
	headerFields := []string{
		"name", "code", "color", "material", "manufacture_office", "manufacture_company", "manufacture_date",
		"us_size", "vn_size", "uk_size", "gender", "attribute", "size_available",
		"qr_code", "rfid_barcode", "price", "currency",
	}

	// Write header row
	if err := writer.Write(headerFields); err != nil {
		return fmt.Errorf("failed to write header to CSV: %w", err)
	}

	// Write data rows
	for _, p := range products {
		record := []string{
			p.Name, p.Code, p.Color, p.Material, p.ManufactureOffice, p.ManuFactureCompany, p.ManufactureDate,
			p.USSize, p.VNSize, p.UKSize, p.Gender, p.Attribute, p.SizeAvailable,
			p.QrCode, p.RfidBarcode, p.Price, p.Currency,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record to CSV: %w", err)
		}
	}

	return nil
}

// callBartenderPrinterAPI calls the Bartender Printer API with optimized client
func (ks *KafkaService) callBartenderPrinterAPI(filename string, isFakeCallApi bool, documentFilePath string, connectionSetupPath string) (*model.BartenderApIResponse, error) {
	if isFakeCallApi {
		resp := &http.Response{}
		if rand.Intn(2) == 0 {
			resp = facker.FakeAPICallRunning()
		} else {
			resp = facker.FakeAPICallWaitingToRun()
		}

		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("Fake API call failed with status: %s", resp.Status)
		}

		body, _ := io.ReadAll(resp.Body)
		fmt.Println("========================================API Bartender Printer - RESPONSE====================================================")
		ks.logger.Infof("API response: %s\n", string(body))
		fmt.Println("========================================API Bartender Printer - RESPONSE====================================================")

		bartenderResponse := &model.BartenderApIResponse{}
		if err := json.Unmarshal(body, &bartenderResponse); err != nil {
			ks.logger.Errorf("JSON unmarshal error at Bartender API response: %v", err)
			return nil, nil
		}

		return bartenderResponse, nil
	}

	url := ks.config.BartenderPrinterAPI.URL
	username := ks.config.BartenderPrinterAPI.Username
	password := ks.config.BartenderPrinterAPI.Password

	payload := ""
	if documentFilePath != "" && connectionSetupPath != "" {
		payload = fmt.Sprintf(`ActionGroup:
  Actions:
    - TransformTextToRecordSetAction:
        ConnectionSetup:
          File: D:\hsk-bar\%s
        Text:
          File: D:\hsk-bar\data\%s
        RecordSetVariableName: datum
    - PrintBTWAction:
        DocumentFile: D:\hsk-bar\%s
        Printer: HASAKI-RFID
        SaveAfterPrint: false
        Copies: 1
        DatabaseOverrides:
          - Name: db
            Type: VariableName
            DataSourceVariableName: datum`, connectionSetupPath, filename, documentFilePath)
	} else {
		return nil, errors.New("templatePath is empty, cannot call Bartender Printer API")
	}

	ks.logger.Infof("URL: %s, Payload: %s", url, payload)

	// Use context with timeout
	ctx, cancel := context.WithTimeout(ks.ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, ks.config.BartenderPrinterAPI.Method, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "text/vnd.yaml")
	req.SetBasicAuth(username, password)

	// Use the optimized client with connection pooling
	resp, err := ks.bartenderClient.Do(req)
	if err != nil {
		ks.logger.Errorf("Send request to Bartender Printer API error: %v", err)
		return nil, err
	}
	ks.logger.Infof("Call API Printer Successfully: %s", filename)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			ks.logger.Errorf("Failed to close response body: %s", err)
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)

	bartenderResponse := &model.BartenderApIResponse{}
	// Check response status
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bartender API returned error status: %d, body: %s", resp.StatusCode, string(body))
	} else {
		fmt.Println("========================================API Bartender Printer - RESPONSE====================================================")
		ks.logger.Infof("API response: %s\n", string(body))
		fmt.Printf("API response: %s\n", string(body))
		fmt.Println("========================================API Bartender Printer - RESPONSE====================================================")
		// Parse the response body
		if err := json.Unmarshal(body, &bartenderResponse); err != nil {
			ks.logger.Errorf("JSON unmarshal error at Bartender API response: %v", err)
			return nil, nil
		}
	}

	// Optionally check job status for monitoring
	// ks.checkBartenderPrinterAPIStatus(body)
	return bartenderResponse, nil
}

func (ks *KafkaService) checkBartenderPrinterAPIStatus(body []byte) {
	//BartenderApIResponse: {"Id":"c6027b37-c08e-4284-b36d-d35122fdf798","Status":"Running","StatusUrl":"http://127.0.0.1:5159/api/actions/c6027b37-c08e-4284-b36d-d35122fdf798"}
	var bartenderApIResponse model.BartenderApIResponse
	if err := json.Unmarshal(body, &bartenderApIResponse); err != nil {
		ks.logger.Errorf("JSON unmarshal error: %v", err)
		return
	}

	if bartenderApIResponse.Status == "Running" || bartenderApIResponse.Status == "WaitingToRun" ||
		strings.Contains(bartenderApIResponse.Status, "run") {
		ks.logger.Infof("Bartender Printer API is running: %s", bartenderApIResponse.Status)
		// Call API to check the status of the job
		err := ks.callBartenderPrinterAPIStatus(bartenderApIResponse.StatusUrl)
		if err != nil {
			ks.logger.Errorf("Error checking Bartender Printer API status: %v", err)
			return
		}
	}

}

// checkBartenderPrinterAPIStatus checks the status of the job
func (ks *KafkaService) callBartenderPrinterAPIStatus(statusUrl string) error {
	// http://localhost:5159/api/actions/f39f0ab2-3db8-4ad0-a56a-6f3ba94c0410?MessageCount=200&MessageSeverity=Info&Variables=PrintJobStatus%2CResponse
	buildUrl := statusUrl + "?MessageCount=200&MessageSeverity=Info&Variables=PrintJobStatus%2CResponse"

	req, err := http.NewRequest(ks.config.BartenderTrackingScriptAPI.Method, buildUrl, nil)
	if err != nil {
		return err
	}

	username := ks.config.BartenderTrackingScriptAPI.Username
	password := ks.config.BartenderTrackingScriptAPI.Password
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "text/vnd.yaml")
	req.SetBasicAuth(username, password)

	// Use NTLM authentication
	client := http.Client{
		Transport: ntlmssp.Negotiator{
			RoundTripper: http.DefaultTransport,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		ks.logger.Errorf("Send request tracking script to Bartender Printer API error: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			ks.logger.Errorf("Failed to close response body: %s", err)
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("==============================API Tracking Script - RESPONSE============================")
	ks.logger.Infof("API Tracking Script Response	: %s\n", string(body))
	fmt.Println("==============================API Tracking Script - RESPONSE============================")
	return nil
}

// populateAndRemakeProducts processes products and encodes SizeAvailable
func (ks *KafkaService) populateAndRemakeProducts(products []*model.Product) {
	for _, p := range products {
		if ks.config.IsUsedImgLocalPath {
			fileName := ks.getGenderSizeAvailablePath(p.VNSize, p.Gender)
			p.SizeAvailable = fileName
		} else {
			sizeAvailableBase64, err := ks.base64Encode(p.SizeAvailable)
			if err != nil {
				ks.logger.Errorf("Error encoding SizeAvailable for product %s: %v", p.Name, err)
				continue
			}
			p.SizeAvailable = sizeAvailableBase64
		}
	}
}

// getDocumentFilePath returns the document file path based on template
func (ks *KafkaService) getDocumentFilePath(template string) (string, error) {
	switch strings.ToLower(template) {
	case string(constant.KidVn):
		return "kid_vn\\kid_vn_noprice.btw", nil
	case string(constant.AdultVn):
		return "adult_vn\\adult_vn_noprice.btw", nil
	case string(constant.KidUs):
		return "kid_us\\kid_us_noprice.btw", nil
	case string(constant.AdultUs):
		return "adult_us\\adult_us_noprice.btw", nil
	default:
		return "", errors.Errorf("Invalid template type: %s", template)
	}
}

// getConnectionFilePath returns the connection file path based on template
func (ks *KafkaService) getConnectionFilePath(template string) (string, error) {
	switch strings.ToLower(template) {
	case string(constant.KidVn):
		return "kid_vn\\db.xml", nil
	case string(constant.AdultVn):
		return "adult_vn\\db.xml", nil
	case string(constant.KidUs):
		return "kid_us\\db.xml", nil
	case string(constant.AdultUs):
		return "adult_us\\db.xml", nil
	default:
		return "", errors.Errorf("Invalid template type: %s", template)
	}
}

func (ks *KafkaService) getGenderSizeAvailablePath(vnSize string, gender string) string {
	switch strings.ToLower(gender) {
	case string(constant.GenderTypeWomen), string(constant.GenderTypeMen):
		return ks.getAdultSizeAvailablePath(vnSize)
	case string(constant.GenderTypeKid):
		return ks.getKidSizeAvailablePath(vnSize)
	default:
		return ""
	}
}

func (ks *KafkaService) getAdultSizeAvailablePath(vnSize string) string {
	switch strings.ToLower(vnSize) {
	case string(constant.AdultSizeAvailableXS):
		return "adult_xs.pdf"
	case string(constant.AdultSizeAvailableS):
		return "adult_s.pdf"
	case string(constant.AdultSizeAvailableM):
		return "adult_m.pdf"
	case string(constant.AdultSizeAvailableL):
		return "adult_l.pdf"
	case string(constant.AdultSizeAvailableXL):
		return "adult_xl.pdf"
	case string(constant.AdultSizeAvailable2XL):
		return "adult_2xl.pdf"
	case string(constant.AdultSizeAvailable3XL):
		return "adult_3xl.pdf"
	default:
		return ""
	}
}

func (ks *KafkaService) getKidSizeAvailablePath(vnSize string) string {
	switch strings.ToLower(vnSize) {
	case string(constant.KidSizeAvailable3T):
		return "kids_3t.pdf"
	case string(constant.KidSizeAvailable4T):
		return "kids_4t.pdf"
	case string(constant.KidSizeAvailable5):
		return "kids_5.pdf"
	case string(constant.KidSizeAvailable6):
		return "kids_6.pdf"
	case string(constant.KidSizeAvailable7):
		return "kids_7.pdf"
	case string(constant.KidSizeAvailable8):
		return "kids_8.pdf"
	default:
		return ""
	}
}

// base64Encode fetches data from URL and encodes to base64
func (ks *KafkaService) base64Encode(url string) (string, error) {
	if url == "" {
		return "", errors.Errorf("Product attribute url is empty")
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			ks.logger.Errorf("Failed to close response body: %s", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(body), nil
}
