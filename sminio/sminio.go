package sminio

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/androidsr/sc-go/syaml"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	client     *minio.Client
)

func New(cfg *syaml.MinioInfo) {
	var err error
	client, err = minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalln("创建 MinIO 客户端失败", err)
		return
	}
}

func CreateBucket(bucketName string) bool {
	err := client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// 检查存储桶是否已经存在。
		exists, err := client.BucketExists(context.Background(), bucketName)
		if err == nil && exists {
			return true
		} else {
			log.Printf("创建存储桶失败：%v\n", err)
			return false
		}
	}
	return true
}

// UploadFile 上传文件到指定的 Minio 存储桶中。
func UploadFile(client *minio.Client, bucketName, objectName, filePath, contentType string) error {
	// 打开本地文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err = client.PutObject(ctx, bucketName, objectName, file, fileInfo.Size(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	fmt.Printf("成功上传文件 %s 到存储桶 %s 中\n", objectName, bucketName)
	return nil
}

// DownloadFile 从Minio存储桶中下载文件到本地。
func DownloadFile(bucketName, objectName, filePath string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader, err := client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}

// ListObjects 列出指定存储桶中的对象。
func ListObjects(bucketName string) ([]string, error) {
	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用ListObjects列出存储桶中的对象
	objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{WithMetadata: true})
	var objectNames []string

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		objectNames = append(objectNames, object.Key)
	}

	return objectNames, nil
}

// RemoveObject 从Minio存储桶中删除指定的对象。
func RemoveObject(bucketName, objectName string) error {
	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用RemoveObject删除对象
	err := client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}
