# Google Cloud Platform Storage Bucket helper





## Installation

```sh
go get -u github.com/ahmadissa/gcp_storage
```


## Example

you need to export GOOGLE_APPLICATION_CREDENTIALS

for more information how to get service account json key file check:


https://cloud.google.com/iam/docs/creating-managing-service-account-keys

```
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
```

to run example

```

cd example
go run example.go

```

```go

package main

import (
	"fmt"
	"time"

	"github.com/ahmadissa/gcp_storage"
)

func main() {
	GCPStorage.Init("your_bucket_name")
	localFile := "../testFiles/localfile.txt"
	cloudFile := "test.txt"

	//upload file
	err := GCPStorage.Upload(localFile, cloudFile)
	if err != nil {
		//handle error
		panic(err)
	}

	//get file siza
	size, err := GCPStorage.Size(cloudFile)
	if err != nil {
		//handle error
		panic(err)
	}
	fmt.Printf("size in Bytes: %v\n", size)

	//get file md5 hash
	hash, err := GCPStorage.MD5(cloudFile)
	if err != nil {
		//handle error
		panic(err)
	}
	fmt.Printf("md5 hash: %v\n", hash)

	//check if file exists
	exists, _ := GCPStorage.Exists(cloudFile)
	fmt.Printf("file exists: %v\n", exists)

	//download file
	err = GCPStorage.Download(cloudFile, "localFile.txt")
	if err != nil {
		//handle error
		panic(err)
	}

	//get all file meta data
	attrs, err := GCPStorage.Attrs(cloudFile)
	if err != nil {
		//handle error
		panic(err)
	}
	fmt.Printf("file attrs: %v\n", attrs)

	//list all files
	files, err := GCPStorage.List("")
	if err != nil {
		//handle error
		panic(err)
	}
	fmt.Printf("files: %v\n", files)
	//make public and get download url
	// if you want to test the download url make sure you dont delete the file in last example function
	url, err := GCPStorage.MakePublic(cloudFile)
	if err != nil {
		//handle error
		panic(err)
	}
	fmt.Printf("download url: %v\n", url)
	//delete file
	err = GCPStorage.Delete(cloudFile)
	if err != nil {
		//handle error
		panic(err)
	}

	//Delete folder
	err = GCPStorage.DeleteFolder("instagram/cache")
	if err != nil {
		//handle error
		panic(err)
	}
	//Delete files inside a folder which is one hour old or more
	err = GCPStorage.DeleteOldFiles("instagram/", time.Hour)
	if err != nil {
		//handle error
		panic(err)
	}
}



```
## Test

```
go test
```
## License

GNU GENERAL PUBLIC LICENSE. See the LICENSE file for details.