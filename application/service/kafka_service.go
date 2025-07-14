package service

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Azure/go-ntlmssp"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/pkg/errors"

	"kafka-consumer/application/constant"
	"kafka-consumer/application/logger"
	"kafka-consumer/application/model"
	"kafka-consumer/config"
)

// KafkaService contains all dependencies needed for Kafka operations
type KafkaService struct {
	logger logger.ILogger
	config *config.Config
}

// NewKafkaService creates a new KafkaService instance
func NewKafkaService(logger logger.ILogger, config *config.Config) *KafkaService {
	return &KafkaService{
		logger: logger,
		config: config,
	}
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
	now := time.Now()
	fileFormat := constant.FileTypeTxt
	filename := "test" + "_" + now.Format("20060102_150405") + "." + string(fileFormat)
	filepath := ks.config.FileSharePath + string(os.PathSeparator) + filename

	ks.logger.Infof("Export file to path successfully!: %s", filepath)

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

	ks.populateAndRemakeProducts(productPrinterMsg.Products)

	if err := ks.exportProducts(productPrinterMsg.Products, filepath, fileFormat); err != nil {
		ks.logger.Errorf("Export error: %v", err)
		return err
	}

	ks.logger.Infof("Exported products to %s at time: %s", filepath, now)

	if err := ks.bartenderPrinterAPI(filename, true, documentFilePath, connectionSetupFile); err != nil {
		ks.logger.Errorf("API call error: %v", err)
		return err
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

// bartenderPrinterAPI calls the Bartender Printer API
func (ks *KafkaService) bartenderPrinterAPI(filename string, isCallAPI bool, documentFilePath string, connectionSetupPath string) error {
	ks.logger.Infof("Call API Printer Successully: %s", filename)

	if !isCallAPI {
		return nil
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
		return errors.New("templatePath is empty, cannot call Bartender Printer API")
	}

	fmt.Println("===================================URL - PAYLOAD===================================================")
	ks.logger.Infof("URL: %s, Payload: %s\n", url, payload)
	fmt.Println("===================================================================================================")
	req, err := http.NewRequest(ks.config.BartenderPrinterAPI.Method, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
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
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			ks.logger.Errorf("Failed to close response body: %s", err)
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	ks.logger.Infof("API response: %s\n", string(body))
	fmt.Println("========================================API - RESPONSE====================================================")
	fmt.Println("API response:", string(body))
	fmt.Println("========================================API - RESPONSE====================================================")

	return nil
}

// populateAndRemakeProducts processes products and encodes SizeAvailable
func (ks *KafkaService) populateAndRemakeProducts(products []*model.Product) {
	for _, p := range products {
		sizeAvailableBase64, err := ks.base64Encode(p.SizeAvailable)
		if err != nil {
			ks.logger.Errorf("Error encoding SizeAvailable for product %s: %v", p.Name, err)
			continue
		}
		p.SizeAvailable = sizeAvailableBase64
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
