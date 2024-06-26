package fs_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lks-go/url-shortener/pkg/fs"
)

func ExampleProducer_WriteRow() {
	defer os.Remove("./example_write_row")

	testRecords := []fs.Record{
		{
			UUID:        "test1",
			ShortURL:    "x",
			OriginalURL: "xxx",
		},
		{
			UUID:        "test2",
			ShortURL:    "x",
			OriginalURL: "xxx",
		},
		{
			UUID:        "test3",
			ShortURL:    "x",
			OriginalURL: "xxx",
		},
	}

	c, err := fs.NewProducer("./example_write_row")
	if err != nil {
		panic(err)
	}

	for _, rec := range testRecords {
		err := c.WriteRow(&rec)
		if err != nil {
			panic(err)
		}
	}
	c.Close()

	f, err := os.OpenFile("./example_write_row", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	recordsInFile := make([]fs.Record, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		rec := fs.Record{}
		err = json.Unmarshal(scanner.Bytes(), &rec)
		if err != nil {
			panic(err)
		}

		recordsInFile = append(recordsInFile, rec)
	}

	fmt.Println(recordsInFile)

	// Output:
	// [{test1 x xxx } {test2 x xxx } {test3 x xxx }]
}

func ExampleConsumer_ReadRow() {
	defer os.Remove("./example_read_row")

	testRecords := fs.Record{
		UUID:        "test1",
		ShortURL:    "x",
		OriginalURL: "xxx",
	}

	f, err := os.OpenFile("./example_read_row", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(testRecords)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(b)
	if err != nil {
		panic(err)
	}

	_, err = f.Write([]byte("\n"))
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	c, err := fs.NewConsumer("./example_read_row")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	gotRec := fs.Record{}
	err = c.ReadRow(&gotRec)
	if err != nil {
		panic(err)
	}

	fmt.Println(gotRec)

	// Output:
	// {test1 x xxx }
}
