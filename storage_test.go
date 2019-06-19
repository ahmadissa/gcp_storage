package GCPStorage

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	bucketName := os.Getenv("GOOGLE_CLOUD_BUCKET")
	if bucketName == "" {
		t.Error("Environment variable 'GOOGLE_CLOUD_BUCKET' was not set")
	}
	Init(bucketName)
}

func TestUploadDeleteExists(t *testing.T) {
	src := "testFiles/localfile.txt"
	dst := "tempFile.txt"

	err := Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	exists, err := Exists(dst)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Error("File does not exists")
	}
	err = Delete(dst)
	if err != nil {
		t.Error(err)
	}
	exists, _ = Exists(dst)

	if exists {
		t.Error("File should not exists")
	}
}

func TestDownload(t *testing.T) {
	src := "testFiles/localfile.txt"
	dst := "tempFile.txt"
	temp := "./tempFile.txt"

	err := Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	err = Download(dst, temp)
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

	err := Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = Delete(dst)
		if err != nil {
			t.Error(err)
		}
	}()
	md5, err := MD5(dst)
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

	err := Upload(src, dst)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = Delete(dst)
		if err != nil {
			t.Error(err)
		}
	}()
	size, err := Size(dst)
	if err != nil {
		t.Error(err)
	}
	if size != 9 {
		t.Errorf("size didnt match expecting 9, got: %v", size)
	}

}
