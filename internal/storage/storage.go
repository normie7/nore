package storage

import (
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"path"
	"syscall"

	"github.com/normie7/nore/internal/noiseremover"
)

type storage struct {
	uploadedFilesDirectory  string
	completedFilesDirectory string
	thresholdMb             int64
}

func NewFileStorage(thresholdMb int64, webDir string) noiseremover.Storage {
	return &storage{
		thresholdMb:             thresholdMb,
		uploadedFilesDirectory:  path.Join(webDir, "/temp-files/up"),
		completedFilesDirectory: path.Join(webDir, "/temp-files/ready"),
	}
}

func (s *storage) OpenFileUpFolder(filename string) (*os.File, error) {
	return os.Open(path.Join(s.uploadedFilesDirectory, filename))
}

func (s *storage) RemoveFileUpFolder(filename string) error {
	return os.Remove(path.Join(s.uploadedFilesDirectory, filename))
}

func (s *storage) CreateFileReadyFolder(filename string) (*os.File, error) {
	return os.Create(path.Join(s.completedFilesDirectory, filename))
}

func (s *storage) OpenFileReadyFolder(filename string) (*os.File, error) {
	return os.Open(path.Join(s.completedFilesDirectory, filename))
}

func (s *storage) RemoveFileReadyFolder(filename string) error {
	return os.Remove(path.Join(s.completedFilesDirectory, filename))
}

func (s *storage) IsEnoughSpaceLeft() (bool, error) {
	availableBytes, err := diskUsage(s.uploadedFilesDirectory)
	if err != nil {
		return false, err
	}
	/*
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
	*/
	return availableBytes/(1024*1024) > uint64(s.thresholdMb), nil
}

// disk usage of path/disk
// returns available space in bytes
func diskUsage(path string) (availableBytes uint64, err error) {
	fs := syscall.Statfs_t{}
	err = syscall.Statfs(path, &fs)
	if err != nil {
		return
	}

	return fs.Bavail * uint64(fs.Bsize), nil
}

func (s *storage) Store(file multipart.File) (fileName string, err error) {
	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile(s.uploadedFilesDirectory, "upload-*.mp3")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	// write this byte array to our temporary file
	_, err = tempFile.Write(fileBytes)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), err
}

func (s *storage) Find(fileName string) (*noiseremover.DownloadData, error) {
	p := path.Join(s.completedFilesDirectory, fileName)

	if _, err := os.Stat(p); err == nil {
		return &noiseremover.DownloadData{FullPath: p}, nil
	} else {
		// todo separate file not found log
		log.Println(err)
		return &noiseremover.DownloadData{}, err
	}
}
