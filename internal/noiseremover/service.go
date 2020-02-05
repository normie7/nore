package noiseremover

import (
	"context"
	"mime/multipart"
	"os"
)

const (
	LogToken = "token"

	ProgressNew              Progress = "new"
	ProgressQueued           Progress = "queued"
	ProgressInProgress       Progress = "in_progress"
	ProgressCompleted        Progress = "completed"
	ProgressErrorEncountered Progress = "error_encountered"
)

type Storage interface {
	IsEnoughSpaceLeft() (bool, error)
	Store(file multipart.File) (fileName string, err error)
	Find(fileName string) (*DownloadData, error)
	OpenFileUpFolder(filename string) (*os.File, error)
	RemoveFileUpFolder(filename string) error
	CreateFileReadyFolder(filename string) (*os.File, error)
	OpenFileReadyFolder(filename string) (*os.File, error)
	RemoveFileReadyFolder(filename string) error
}

type Repository interface {
	Add(file *File) error
	SetProgress(fileId string, progress Progress) error
	GetInfo(fileId string) (*File, error)
	GetFilesToProcess() ([]File, error)
	QueueFiles(counter int64) ([]File, error)
	Close() error
}

type Service interface {
	Store(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (*File, error)
	Status(ctx context.Context, fileId string) (*StatusData, error)
	Find(ctx context.Context, fileId string) (*DownloadData, error)
}

type Progress string

type File struct {
	Id           string
	InternalName string
	UploadedName string
	Progress     Progress
}

// will change with cloud storage
type DownloadData struct {
	FullPath string
}

type noiseRemoverService struct {
	storage Storage
	repo    Repository
}

func NewNoiseRemoverService(storage Storage, repo Repository) Service {
	return &noiseRemoverService{
		storage,
		repo,
	}
}
