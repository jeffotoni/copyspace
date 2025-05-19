package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// Mock S3 Client implementing s3iface.S3API
type MockS3Client struct {
	s3iface.S3API
	PutObjectFunc          func(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	GetObjectFunc          func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	ListObjectsV2PagesFunc func(*s3.ListObjectsV2Input, func(*s3.ListObjectsV2Output, bool) bool) error
}

func (m *MockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return m.PutObjectFunc(input)
}
func (m *MockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return m.GetObjectFunc(input)
}
func (m *MockS3Client) ListObjectsV2Pages(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
	return m.ListObjectsV2PagesFunc(input, fn)
}

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	if !IsDir(tmpDir) {
		t.Errorf("Expected %s to be a directory", tmpDir)
	}
	if IsDir(tmpFile) {
		t.Errorf("Expected %s to be a file, not directory", tmpFile)
	}
	if IsDir("nonexistentpath") {
		t.Errorf("Expected nonexistent path to not be a directory")
	}
}

func TestGetFileContentType(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	content := []byte("Hello, World!")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	f, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	contentType, err := GetFileContentType(f)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if contentType != "text/plain; charset=utf-8" && contentType != "application/octet-stream" {
		t.Errorf("Expected content type to be 'text/plain; charset=utf-8' or 'application/octet-stream', got %q", contentType)
	}
}

func TestSendFileDO(t *testing.T) {
	mockClient := &MockS3Client{
		PutObjectFunc: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
			if input == nil || input.Body == nil {
				t.Errorf("PutObjectInput or Body is nil")
			}
			// simulate success
			return &s3.PutObjectOutput{
				ETag: aws.String("mock-etag"),
			}, nil
		},
	}

	// Create temp file to send
	tmpFile, err := os.CreateTemp("", "testupload")
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("test content")
	tmpFile.Write(content)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	job := sendS3{
		Path:     tmpFile.Name(),
		Pbucket:  "test/key.txt",
		S3Client: mockClient, // agora funciona!
		Counter:  1,
	}

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SendFileDO(job)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "mock-etag") {
		t.Errorf("Expected output to contain mock-etag, got: %s", out)
	}
}

func TestDownloadAllObjects(t *testing.T) {
	mockClient := &MockS3Client{
		ListObjectsV2PagesFunc: func(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
			page := &s3.ListObjectsV2Output{
				Contents: []*s3.Object{
					{Key: aws.String("file1.txt")},
					{Key: aws.String("dir1/")},
					{Key: aws.String("dir1/file2.txt")},
				},
			}
			fn(page, true)
			return nil
		},
		GetObjectFunc: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			content := "test data"
			return &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader(content)),
			}, nil
		},
	}

	tmpDir := t.TempDir()
	err := DownloadAllObjects(mockClient, "mock-bucket", tmpDir)
	if err != nil {
		t.Errorf("DownloadAllObjects failed: %v", err)
	}

	// Check if files were created
	file1 := filepath.Join(tmpDir, "file1.txt")
	if _, err := os.Stat(file1); err != nil {
		t.Errorf("Expected file %s to be created", file1)
	}

	file2 := filepath.Join(tmpDir, "dir1", "file2.txt")
	if _, err := os.Stat(file2); err != nil {
		t.Errorf("Expected file %s to be created", file2)
	}
}
