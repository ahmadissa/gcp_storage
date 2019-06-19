package GCPStorage

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/google-api-go-client/iterator"
	"google.golang.org/api/option"
	raw "google.golang.org/api/storage/v1"
)

//export GOOGLE_APPLICATION_CREDENTIALS="/home/user/Downloads/[FILE_NAME].json"

var bucketName string

//Init storage instance
func Init(bucket string) {
	bucketName = bucket
}

//Upload local file to the current bucket
func Upload(localFile, dst string) error {
	f, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer f.Close()
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)

	wc := bucket.Object(dst).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}
	return wc.Close()
}

//Exists check if file exists
func Exists(filePath string) (bool, error) {
	md5, err := MD5(filePath)
	if err != nil {
		return false, err
	}
	if md5 != "" {
		return true, nil
	}
	return false, errors.New("could not get checksum of the file")
}

//Delete storage file from the current bucket
func Delete(filePath string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	return bucket.Object(filePath).Delete(ctx)
}

//Download file from source (src) to local destination (dst)
func Download(src, dst string) error {
	ctx := context.Background()
	// get readonly client
	client, err := storage.NewClient(ctx, option.WithScopes(raw.DevstorageReadOnlyScope))
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	rc, err := bucket.Object(src).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, rc)
	if err != nil {
		return err
	}
	return nil
}

//Attrs returns the metadata for the bucket.
func Attrs(filePath string) (attrs *storage.ObjectAttrs, err error) {
	ctx := context.Background()
	// get readonly client
	client, err := storage.NewClient(ctx, option.WithScopes(raw.DevstorageReadOnlyScope))
	if err != nil {
		return
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	// [START get_metadata]
	o := bucket.Object(filePath)
	return o.Attrs(ctx)
}

//DeleteOldFiles delete files from folder based on their age, time from created date
func DeleteOldFiles(folder string, fileAge time.Duration) error {
	ctx := context.Background()
	// get readonly client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	it := bucket.Objects(ctx, &storage.Query{
		Prefix: folder,
	})
	now := time.Now()
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if diff := now.Sub(attrs.Created); diff > fileAge {
			err = bucket.Object(attrs.Name).Delete(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//Size get the size of the file in int64
func Size(filePath string) (size int64, err error) {
	attrs, err := Attrs(filePath)
	if err != nil {
		return
	}
	size = attrs.Size
	return
}

//MakePublic make file public (readonly) and retrive the download url
func MakePublic(filePath string) (downloadURL string, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	acl := bucket.Object(filePath).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", err
	}
	return "https://storage.googleapis.com/" + bucketName + "/" + filePath, nil
}

//MD5 get the md5 checksum of a file in a bucket
func MD5(filePath string) (md5String string, err error) {
	attrs, err := Attrs(filePath)
	if err != nil {
		return
	}
	md5String = hex.EncodeToString(attrs.MD5[:])
	if md5String != "" {
		return
	}
	err = errors.New("could not get md5 of the file")
	return
}
