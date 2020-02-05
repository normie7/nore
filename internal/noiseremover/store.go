package noiseremover

import (
	"context"
	"log"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
)

type ExtMime struct {
	Ext  string
	Mime string
}

var allowedExt = []ExtMime{
	{
		Ext:  ".mp3",
		Mime: "audio/mpeg",
	},
	{
		Ext:  ".mp3",
		Mime: "audio/mp3",
	},
}

func (n *noiseRemoverService) Store(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*File, error) {
	token := ctx.Value(LogToken)

	log.Printf("%s saving file: %+v, File Size: %+v, MIME Header: %+v\n", token, header.Filename, header.Size, header.Header)
	fileValid := isFileValid(header)
	if !fileValid {
		return nil, ErrWrongFileType
	}

	enoughSpaceLeft, err := n.storage.IsEnoughSpaceLeft()
	if err != nil {
		return nil, err
	}

	if !enoughSpaceLeft {
		return nil, ErrNoSpaceLeft
	}

	name, err := n.storage.Store(file)
	if err != nil {
		return nil, err
	}

	f := &File{
		Id:           uuid.New().String(),
		InternalName: filepath.Base(name),
		UploadedName: header.Filename,
		Progress:     ProgressNew,
	}

	err = n.repo.Add(f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func isFileValid(header *multipart.FileHeader) bool {

	ext := filepath.Ext(header.Filename)
	for _, e := range allowedExt {
		if e.Ext == ext && e.Mime == header.Header.Get("Content-Type") {
			return true
		}
	}
	return false
}
