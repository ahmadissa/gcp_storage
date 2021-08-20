package GCPStorage

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	humanize "github.com/dustin/go-humanize"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	raw "google.golang.org/api/storage/v1"
)

//Meta holds important meta about a file
type Meta struct {
	MD5        string
	Size       int64
	SizeStr    string
	LastUpdate time.Time
	Created    time.Time
}

//export GOOGLE_APPLICATION_CREDENTIALS="/home/user/Downloads/[FILE_NAME].json"

var bucketName string

//Init storage instance
func Init(bucket string) {
	bucketName = bucket
}

//CopyFolder copy cloud storage folder to another dst
func CopyFolder(srcFolder, dstFolder string, multiple bool) error {
	ctx := context.Background()
	// get readonly client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	it := bucket.Objects(ctx, &storage.Query{
		Prefix: srcFolder,
	})
	wg := sync.WaitGroup{}
	errs := []error{}
	for {

		attrs, err := it.Next()
		if err != nil {
			break
		}

		pat := regexp.MustCompile("^(.*?)" + srcFolder + "(.*)$")
		repl := "${1}" + dstFolder + "$2"

		dst := pat.ReplaceAllString(attrs.Name, repl)
		if multiple {
			wg.Add(1)
			go func() {
				err = CopyFile(attrs.Name, dst)
				errs = append(errs, err)
				wg.Done()
			}()

		} else {
			err = CopyFile(attrs.Name, dst)
			if err != nil {
				return err
			}
		}

	}
	if multiple {
		wg.Wait()
		for i := range errs {
			if errs[i] != nil {
				return err
			}
		}
	}
	return err
}

//CopyFile copy cloud storage file to another dst
func CopyFile(src, dst string) error {

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	srcFile := bucket.Object(src)
	dstFile := bucket.Object(dst)
	// Just copy content.
	_, err = dstFile.CopierFrom(srcFile).Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

//UploadFromReader upload from reader to GCP file
func UploadFromReader(reader io.Reader, dst string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	wc := bucket.Object(dst).NewWriter(ctx)
	if _, err = io.Copy(wc, reader); err != nil {
		return err
	}
	return wc.Close()
}

//Upload local file to the current bucket
func Upload(localFile, dst string) error {
	fileReader, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer fileReader.Close()
	return UploadFromReader(fileReader, dst)
}

//GetMeta get size
func GetMeta(src string) (Meta, error) {
	meta := Meta{}
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return meta, err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	attrs, err := bucket.Object(src).Attrs(ctx)
	if err != nil {
		//log.Println(err)
		return meta, err
	}
	meta.MD5 = base64.StdEncoding.EncodeToString(attrs.MD5)
	meta.Size = attrs.Size
	meta.SizeStr = humanize.Bytes(uint64(meta.Size))
	meta.Created = attrs.Created
	return meta, nil
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

//List all files in a bucket with a prefix
//prefix can be a folder, if prefix is empty string the function will return all files in the bucket
//limit is number of files to retrive, 0 means all
func List(prefix string, limit int) (files []string, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithScopes(raw.DevstorageReadOnlyScope))
	if err != nil {
		return nil, err
	}
	q := &storage.Query{
		Prefix: prefix,
	}
	if prefix == "" {
		q = nil
	}
	files = []string{}
	it := client.Bucket(bucketName).Objects(ctx, q)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}
		files = append(files, attrs.Name)
		if len(files) > limit && limit > 0 {
			break
		}
	}
	return

}

func read(client *storage.Client, bucket, object string) ([]byte, error) {
	ctx := context.Background()
	// [START download_file]
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
	// [END download_file]
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

//ReadFile into object
func ReadFile(filepath string, obj interface{}) error {
	ctx := context.Background()
	// get readonly client
	client, err := storage.NewClient(ctx, option.WithScopes(raw.DevstorageReadOnlyScope))
	if err != nil {
		return err
	}
	defer client.Close()
	data, err := read(client, bucketName, filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
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

//DeleteFolder delete all files under folder
func DeleteFolder(folder string) error {
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
	for {
		attrs, err := it.Next()
		if err != nil {
			return nil
		}

		err = bucket.Object(attrs.Name).Delete(ctx)
		if err != nil {
			return err
		}
	}
	return nil
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
