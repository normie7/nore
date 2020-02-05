package noiseremover

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
)

type noiseRemoverLoggingService struct {
	Service
}

func NewNoiseRemoverLoggingService(service Service) Service {
	return &noiseRemoverLoggingService{Service: service}
}

func (s *noiseRemoverLoggingService) Store(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*File, error) {
	f, err := s.Service.Store(ctx, file, header)
	if err != nil {
		err = fmt.Errorf("%s service.Store %w", ctx.Value(LogToken), err)
		log.Println(err)
	}
	return f, err
}

func (s *noiseRemoverLoggingService) Status(ctx context.Context, fileId string) (*StatusData, error) {
	sd, err := s.Service.Status(ctx, fileId)
	if err != nil {
		err = fmt.Errorf("%s service.Status %w", ctx.Value(LogToken), err)
		log.Println(err)
	}
	return sd, err
}

func (s *noiseRemoverLoggingService) Find(ctx context.Context, fileId string) (*DownloadData, error) {
	d, err := s.Service.Find(ctx, fileId)
	if err != nil {
		err = fmt.Errorf("%s service.Find %w", ctx.Value(LogToken), err)
		log.Println(err)
	}
	return d, err
}
