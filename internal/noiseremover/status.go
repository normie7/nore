package noiseremover

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// https://blog.golang.org/go1.13-errors

var (
	ErrSomethingWentWrong = errors.New("Something went wrong")

	ErrWrongTokenFormat = errors.New("Wrong token format")
	ErrFileNotFound     = errors.New("file not found")

	ErrWrongFileType = errors.New("wrong file type")
	ErrNoSpaceLeft   = errors.New("no space left in storage")
	ErrProcessError  = errors.New("ProcessFile error")
)

type StatusData struct {
	FileId       string
	UploadedName string
	Progress     Progress
	// constants for template
	ProgressNew              Progress
	ProgressInProgress       Progress
	ProgressCompleted        Progress
	ProgressErrorEncountered Progress
}

func NewStatusData(fileId, uploadedName string, progress Progress) *StatusData {
	return &StatusData{
		FileId:       fileId,
		UploadedName: uploadedName,
		Progress:     progress,

		ProgressNew:              ProgressNew,
		ProgressInProgress:       ProgressInProgress,
		ProgressCompleted:        ProgressCompleted,
		ProgressErrorEncountered: ProgressErrorEncountered,
	}
}

func (n *noiseRemoverService) Status(ctx context.Context, fileId string) (*StatusData, error) {

	if _, err := uuid.Parse(fileId); err != nil {
		return NewStatusData("", "", ""), ErrWrongTokenFormat
	}

	f, err := n.repo.GetInfo(ctx, fileId)
	if err != nil {
		return NewStatusData("", "", ""), err
	}

	return NewStatusData(f.Id, f.UploadedName, f.Progress), nil
}

func (n *noiseRemoverService) Find(ctx context.Context, fileId string) (*DownloadData, error) {

	if _, err := uuid.Parse(fileId); err != nil {
		return &DownloadData{}, ErrWrongTokenFormat
	}

	f, err := n.repo.GetInfo(ctx, fileId)
	if err != nil {
		return &DownloadData{}, err
	}

	// todo separate error if file processing or error encountered
	if f.Progress != ProgressCompleted {
		return &DownloadData{}, ErrFileNotFound
	}

	// todo possible error handling
	return n.storage.Find(f.InternalName)
}
