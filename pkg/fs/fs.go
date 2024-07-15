package fs

import (
	"encoding/json"
	"os"
)

// Record is a struct helps handle file records
type Record struct {
	UUID        string `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

// NewProducer returns a new instance of Producer
func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Producer is a handler of writing data to file
type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

// WriteRow encodes Record to string and stores it to the files
func (p *Producer) WriteRow(r *Record) error {
	return p.encoder.Encode(r)
}

// Close wrapper of file closer
func (p *Producer) Close() error {
	return p.file.Close()
}

// NewConsumer returns a new instance of Consumer
func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

// Consumer is a handler of reading data to file
type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

// ReadRow reads a new Record from file every call
// returns oi.EOF when file ends
func (c *Consumer) ReadRow(r *Record) error {
	if err := c.decoder.Decode(&r); err != nil {
		return err
	}

	return nil
}

// Close wrapper of file closer
func (c *Consumer) Close() error {
	return c.file.Close()
}
