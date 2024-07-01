package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jeffotoni/gcolor"
)

var (
	BUCKET   = ""
	WORKER   = 500 // quantidade de workers trabalhando simultaneamente
	ACL_S3   = "private"
	HOME_DIR = ""
)

type DOKey struct {
	Key      string `json:"key"`
	Secret   string `json:"secret"`
	Endpoint string `json:"endpoint"`
	Region   string `json:"region"`
	Bucket   string `json:"bucket"`
}

func init() {
	user, err := user.Current()
	if err != nil {
		return
	}
	HOME_DIR = user.HomeDir
}

func main() {
	// accessKey := "YOUR_ACCESS_KEY"
	// secretKey := "YOUR_SECRET_KEY"
	// spaceName := "YOUR_SPACE_NAME"
	// endpoint := "https://nyc3.digitaloceanspaces.com" // Altere para o endpoint correto
	// filePath := "path/to/your/large/file" // Altere para o caminho do seu arquivo

	key, err := ReadKey()
	if err != nil {
		fmt.Println("err:", err)
		fmt.Println("Erro ao montar suas credenciais de acesso ao DigitalOcean Space!")
		return
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(key.Region),
		Endpoint:    aws.String(key.Endpoint),
		Credentials: credentials.NewStaticCredentials(key.Key, key.Secret, ""),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		os.Exit(0)
	}

	////////////
	// agora capturando dados..
	var pathFile string
	var workers int

	flag.StringVar(&pathFile, "file", "", "nome do arquivo ou diretorio a ser enviado")
	aclSend := flag.String("acl", "private", "permissao: public or private")
	fbucket := flag.String("bucket", "", "o nome do seu bucket")
	flag.IntVar(&workers, "worker", WORKER, "quantidade de trabalhos concorrentes em sua máquina")
	flag.Parse()

	if len(pathFile) == 0 {
		flag.PrintDefaults()
		return
	}

	if len(*aclSend) > 0 &&
		strings.ToLower(*aclSend) != "private" &&
		strings.ToLower(*aclSend) != "public" {
		flag.PrintDefaults()
		return
	}

	if len(*aclSend) > 0 && strings.ToLower(*aclSend) == "public" {
		ACL_S3 = "public-read-write"
	}

	if len(*fbucket) > 0 {
		BUCKET = *fbucket
	} else {
		BUCKET = key.Bucket
	}

	if workers > 0 {
		WORKER = workers
	}

	// send file origin
	println(gcolor.CyanCor("domain: " + key.Endpoint))
	println(gcolor.CyanCor("domain copy: https://" + BUCKET + "." + strings.Replace(key.Endpoint, "https://", "", -1)))
	println(gcolor.YellowCor("bucket: " + BUCKET))
	println(gcolor.YellowCor("acl: " + ACL_S3))
	println(gcolor.YellowCor("path: " + pathFile))

	if IsDir(pathFile) {
		fmt.Println("Error, necessário que seja arquivo:", err)
		os.Exit(0)
	}

	pbucket := strings.Replace(pathFile, HOME_DIR, "", -1)

	////////////

	svc := s3.New(sess)

	// Open the file for reading
	file, err := os.Open(pathFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create the multipart upload
	createResp, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(pbucket),
		Key:    aws.String(pathFile),
	})
	if err != nil {
		fmt.Println("Error creating multipart upload:", err)
		os.Exit(1)
	}

	uploadID := createResp.UploadId
	partNum := int64(1)
	var completedParts []*s3.CompletedPart

	// Upload the file in parts
	buf := make([]byte, 5*1024*1024) // 5 MB buffer
	for {
		n, err := file.Read(buf)
		if err != nil {
			break
		}

		partResp, err := svc.UploadPart(&s3.UploadPartInput{
			Bucket:     aws.String(pbucket),
			Key:        aws.String(pathFile),
			PartNumber: aws.Int64(partNum),
			UploadId:   uploadID,
			Body:       bytes.NewReader(buf[:n]),
		})
		if err != nil {
			fmt.Println("Error uploading part:", err)
			os.Exit(1)
		}

		completedParts = append(completedParts, &s3.CompletedPart{
			ETag:       partResp.ETag,
			PartNumber: aws.Int64(partNum),
		})

		partNum++
	}

	// Complete the multipart upload
	_, err = svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(pbucket),
		Key:      aws.String(pathFile),
		UploadId: uploadID,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		fmt.Println("Error completing multipart upload:", err)
		os.Exit(1)
	}

	fmt.Println("Upload completed successfully")
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

func ReadKey() (*DOKey, error) {

	cfgFile := fmt.Sprintf("%s/.dokeys", HOME_DIR)
	fmt.Println("cfgFile:", cfgFile)
	b, err := os.ReadFile(cfgFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	key := &DOKey{}
	if err := json.Unmarshal(b, key); err != nil {
		return nil, err
	}
	return key, nil
}

func GetFileContentType(out *os.File) (string, error) {

	if _, err := out.Seek(0, 0); err != nil {
		return "", err
	}

	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	if _, err := out.Seek(0, 0); err != nil {
		return "", err
	}
	return http.DetectContentType(buffer), nil
}
