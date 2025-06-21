package GCPStorage

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func readURL(httpURL string) (string, error) {
	resp, err := http.Get(httpURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
func getBucket() Bucket {
	bucketName := os.Getenv("GOOGLE_CLOUD_BUCKET")
	if bucketName == "" {
		panic("Environment variable 'GOOGLE_CLOUD_BUCKET' was not set")
	}
	bucket := Bucket{}
	bucket.Init(bucketName)
	return bucket
}

func TestUploadDeleteExists(t *testing.T) {
	src := "testFiles/localfile.txt"
	dst := "tempFile.txt"
	bucket := getBucket()
	err := bucket.Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	exists, err := bucket.Exists(dst)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("File does not exists")
	}
	err = bucket.Delete(dst)
	if err != nil {
		t.Error(err)
	}
	exists, _ = bucket.Exists(dst)

	if exists {
		t.Error("File should not exists")
	}
}

func TestGetSignedURL(t *testing.T) {
	src := "testFiles/localfile.txt"
	dst := "tempFileSignged.txt"
	sourceFile, err := ioutil.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	bucket := getBucket()
	err = bucket.Upload(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	defer bucket.Delete(dst)
	signedURL, err := bucket.GetSignedURL(dst, time.Second*3)
	dstData, err := readURL(signedURL)
	if err != nil {
		t.Fatal(err)
	}
	if string(dstData) != string(sourceFile) {
		t.Fatal("Remote file does not match local file")
	}
	time.Sleep(6 * time.Second)
	dstData, err = readURL(signedURL)
	if err == nil && string(dstData) == string(sourceFile) {
		t.Fatal("Remote file did not expire")
	}
}
func TestDownload(t *testing.T) {
	src := "testFiles/localfile.txt"
	dst := "tempFile.txt"
	temp := "./tempFile.txt"
	bucket := getBucket()
	err := bucket.Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	err = bucket.Download(dst, temp)
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(temp)
	if os.IsNotExist(err) {
		t.Error("cloud not download file")
	}
	os.Remove(temp)
}

func TestMD5(t *testing.T) {
	src := "testFiles/localfile.txt"
	dst := "tempFile.txt"
	bucket := getBucket()
	err := bucket.Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = bucket.Delete(dst)
		if err != nil {
			t.Error(err)
		}
	}()
	md5, err := bucket.MD5(dst)
	if err != nil {
		t.Error(err)
	}
	if md5 != "f20d9f2072bbeb6691c0f9c5099b01f3" {
		t.Error("md5 didnt match expecting f20d9f2072bbeb6691c0f9c5099b01f3, got:" + md5)
	}

}
func TestMD5File(t *testing.T) {
	src := "testFiles/localfile.txt"
	md5, err := MD5file(src)
	if err != nil {
		t.Error(err)
	}
	if md5 != "f20d9f2072bbeb6691c0f9c5099b01f3" {
		t.Error("md5 didnt match expecting f20d9f2072bbeb6691c0f9c5099b01f3, got:" + md5)
	}

}
func TestSize(t *testing.T) {
	src := "testFiles/localfile.txt"
	dst := "tempFile.txt"
	bucket := getBucket()
	err := bucket.Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = bucket.Delete(dst)
		if err != nil {
			t.Error(err)
		}
	}()
	size, err := bucket.Size(dst)
	if err != nil {
		t.Error(err)
	}
	if size != 9 {
		t.Errorf("size didnt match expecting 9, got: %v", size)
	}

}

func TestCopyFolder(t *testing.T) {
	srcFolder := "testFiles"
	dstFolder := "testFiles_dst"
	src := srcFolder + "/localfile.txt"
	dst := dstFolder + "/tempFile.txt"
	bucket := getBucket()
	err := bucket.Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = bucket.DeleteFolder(srcFolder)
		if err != nil {
			t.Error(err)
		}
		err = bucket.DeleteFolder(dstFolder)
		if err != nil {
			t.Error(err)
		}
	}()
	bucket.CopyFolder(srcFolder, dstFolder, true)
}

func TestUploadFromURL(t *testing.T) {
	bucket := getBucket()
	dst := "uploadFromURL/test-file.jpg"
	url := "https://via.placeholder.com/150" // Small image for testing

	err := bucket.UploadFromURL(url, dst)
	if err != nil {
		t.Fatalf("UploadFromURL failed: %v", err)
	}
	defer func() {
		_ = bucket.Delete(dst)
	}()

	// Check if file exists
	exists, err := bucket.Exists(dst)
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Fatal("Uploaded file from URL does not exist in bucket")
	}
}
