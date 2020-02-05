package noiseremover

import (
	"sync"
	"time"
)

type BackgroundService interface {
	ProcessTicker(duration time.Duration, timeToStop chan bool, wg *sync.WaitGroup)
	ProcessFile(file File) error
}

type backgroundService struct {
	storage Storage
	repo    Repository
}

func NewBackGroundService(storage Storage, repo Repository) BackgroundService {
	return &backgroundService{storage: storage, repo: repo}
}
