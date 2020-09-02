package noiseremover

import (
	"context"
	"sync"
	"time"
)

type BackgroundService interface {
	ProcessTicker(ctx context.Context, wg *sync.WaitGroup, duration time.Duration)
	ProcessFile(ctx context.Context, file File) error
}

type backgroundService struct {
	storage Storage
	repo    Repository
}

func NewBackGroundService(storage Storage, repo Repository) BackgroundService {
	return &backgroundService{storage: storage, repo: repo}
}
