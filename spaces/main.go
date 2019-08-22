package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

type sendS3 struct {
	Path     string
	Pbucket  string
	S3Client *s3.S3
	Counter  int
}

// DOKey contem dados para autenticacao na Digital Ocean(acho).
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

	// abir file secrey
	key, err := ReadKey()
	if err != nil {
		fmt.Println("Erro ao montar suas credenciais de acesso ao DigitalOcean Space!")
		return
	}

	// Initialize a client using Spaces
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(key.Key, key.Secret, ""),
		Endpoint:    aws.String(key.Endpoint),
		Region:      aws.String(key.Region), // This is counter intuitive, but it will fail with a non-AWS region name.
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

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
	println(gcolor.YellowCor("bucket: " + BUCKET))

	// send one file
	if !IsDir(pathFile) {
		pbucket := strings.Replace(pathFile, HOME_DIR, "", -1)
		//fmt.Println(SendFileDO(pathFile, pbucket, s3Client, 1))
		SendFileDO(sendS3{
			pathFile,
			pbucket,
			s3Client,
			1,
		})
		return
	}

	jobs := make(chan sendS3)
	//results := make(chan string)
	var i int

	// inicia o worker
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for w := 1; w <= WORKER; w++ {
		wg.Add(1)
		go worker(ctx, &wg, jobs)
	}

	if err := filepath.Walk(pathFile,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			i++
			pbucket := strings.Replace(path, os.Getenv("HOME"), "", -1)
			jobs <- sendS3{
				Path:     path,
				Pbucket:  pbucket,
				S3Client: s3Client,
				Counter:  i,
			}
			return nil
		}); err != nil {
		fmt.Println(err)
	}

	cancel()
	wg.Wait()
	close(jobs)
	println("fim de envio")

	return
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan sendS3) {
	for {
		select {
		case j := <-jobs:
			SendFileDO(j)
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

//func SendFileDO(pf, pbucket string, s3Client *s3.S3, I int) string {
func SendFileDO(job sendS3) {

	if IsDir(job.Path) {
		fmt.Printf(gcolor.YellowCor("Copiando diretorio")+": %s\n", job.Path)
		return
	}

	start := time.Now()
	f, err := os.Open(job.Path)
	if err != nil {
		fmt.Printf("Erro enviando %s: %v\n", job.Path, err)
		return
	}
	defer f.Close()

	// size file...
	fi, err := f.Stat()
	if err != nil {
		fmt.Printf("Erro enviando %s: %v\n", job.Path, err)
		return
	}

	contentType, err := GetFileContentType(f)
	if err != nil {
		fmt.Printf("Erro detectando tipo para o arquivo %s: %v\n", job.Path, err)
		return
	}

	//// Use bufio.NewReader to get a Reader.
	// ... Then use ioutil.ReadAll to read the entire content.
	// reader := bufio.NewReader(f)
	// b, err := ioutil.ReadAll(reader)
	// if err != nil {
	// 	fmt.Println("error ao ler conteudo do arquivo:", err)
	// 	return
	// }

	// if len(string(b)) == 0 {
	// 	println("Error file está vazio..")
	// 	return
	// }

	pathV := strings.Split(job.Path, "/")
	nameFileSpace := pathV[len(pathV)-1]

	// Upload a file to the Space
	object := s3.PutObjectInput{
		ACL:         aws.String(ACL_S3),
		Body:        f, // execlente
		Bucket:      aws.String(BUCKET),
		Key:         aws.String(job.Pbucket),
		ContentType: aws.String(contentType),
	}

	msgs3V, err := job.S3Client.PutObject(&object)
	if err != nil {
		fmt.Printf("Erro enviando %s: %v\n", job.Path, err)
		return
	}

	bmsg, err := json.Marshal(map[string]interface{}{
		"Counter":   job.Counter,
		"Id":        *msgs3V.ETag,
		"File":      job.Pbucket,
		"FileSpace": nameFileSpace,
		"Size":      fi.Size() / 1024,
		"Time":      time.Now().Sub(start),
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(bmsg))
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

func ReadKey() (*DOKey, error) {
	// user, err := user.Current()
	// if err != nil {
	// 	return nil, err
	// }
	cfgFile := fmt.Sprintf("%s/.dokeys", HOME_DIR)
	b, err := ioutil.ReadFile(cfgFile)
	if err != nil {
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
