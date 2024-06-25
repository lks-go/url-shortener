// Пакет предназначен для сохранения и чтения записей из файла
package fs

import (
	"encoding/json"
	"os"
)

// Record стуктура отвечает за формат хранения записи в файле
type Record struct {
	UUID        string `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

// NewProducer конструктор продьюсера
// отвечает за добавление записей в файл
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

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

// WriteRow кодирует структуру Record и добавляет запись в файл
func (p *Producer) WriteRow(r *Record) error {
	return p.encoder.Encode(r)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

// NewConsumer конструктор консьюмера
// отвечает за чтение записей из файла
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

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

// ReadRow при каждом вызове последовательно читает новую строку из файла и декодирует в структуру Record
func (c *Consumer) ReadRow(r *Record) error {
	if err := c.decoder.Decode(&r); err != nil {
		return err
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
