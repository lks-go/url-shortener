package fs_test

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lks-go/url-shortener/pkg/fs"
)

const testFileName = "./test_file"

func deleteFile(t *testing.T) {
	t.Helper()

	require.NoError(t, os.Remove(testFileName))
}

func TestConsumer_Close(t *testing.T) {
	defer deleteFile(t)

	c, err := fs.NewConsumer(testFileName)
	require.NoError(t, err)

	r := fs.Record{}
	assert.Equal(t, io.EOF, c.ReadRow(&r))
	require.NoError(t, c.Close())
}

func TestConsumer_ReadRow(t *testing.T) {
	defer deleteFile(t)

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

	f, err := os.OpenFile(testFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	require.NoError(t, err)

	for _, rec := range testRecords {
		b, err := json.Marshal(rec)
		require.NoError(t, err)

		_, err = f.Write(b)
		require.NoError(t, err)

		_, err = f.Write([]byte("\n"))
		require.NoError(t, err)
	}

	require.NoError(t, f.Close())

	c, err := fs.NewConsumer(testFileName)
	require.NoError(t, err)
	defer c.Close()

	for _, expectedRec := range testRecords {
		gotRec := fs.Record{}
		require.NoError(t, c.ReadRow(&gotRec))
		require.Equal(t, gotRec, expectedRec)
	}

	gotRec := fs.Record{}
	require.EqualError(t, c.ReadRow(&gotRec), io.EOF.Error())
}

func TestProducer_Close(t *testing.T) {
	defer deleteFile(t)

	c, err := fs.NewProducer(testFileName)
	require.NoError(t, err)

	require.NoError(t, c.WriteRow(&fs.Record{}))
	require.NoError(t, c.Close())
}

func TestProducer_WriteRow(t *testing.T) {
	defer deleteFile(t)

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

	c, err := fs.NewProducer(testFileName)
	require.NoError(t, err)

	for _, rec := range testRecords {
		require.NoError(t, c.WriteRow(&rec))
	}

	require.NoError(t, c.Close())

	f, err := os.OpenFile(testFileName, os.O_RDONLY, 0666)
	require.NoError(t, err)

	recordsInFile := make([]fs.Record, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		rec := fs.Record{}
		err = json.Unmarshal(scanner.Bytes(), &rec)
		require.NoError(t, err)

		recordsInFile = append(recordsInFile, rec)
	}

	assert.Equal(t, testRecords, recordsInFile)
}
