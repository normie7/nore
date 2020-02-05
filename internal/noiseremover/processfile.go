package noiseremover

import (
	"bufio"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
	"github.com/viert/go-lame"
)

type result struct {
	fileId string
	err    error
}

func (b *backgroundService) ProcessTicker(duration time.Duration, timeToStop chan bool, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		log.Println("backgroundService.ProcessTicker done")
	}()

	//todo make results pulse
	const numWorkers = 3

	var pwg sync.WaitGroup
	jobs := make(chan File, numWorkers)
	addJobs := func(jobs chan<- File) {
		// skip ticker if jobs queue is not empty
		if len(jobs) != 0 {
			return
		}
		files, err := b.repo.QueueFiles(numWorkers)
		if err != nil {
			log.Println("can't get files to process", err)
			return
		}

		for _, f := range files {
			// can't be blocking as jobs queue was empty
			jobs <- f
		}
	}

	results := make(chan result, numWorkers+cap(jobs))
	for w := 1; w <= numWorkers; w++ {
		go b.processFileWorker(w, jobs, results, timeToStop, &pwg)
		pwg.Add(1)
	}

	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case r := <-results:
			log.Println("file", r.fileId, "err", r.err)
			addJobs(jobs)
		case <-ticker.C:
			addJobs(jobs)
		case <-timeToStop:
			// dequeue all the jobs
			close(jobs)
			for j := range jobs {
				err := b.repo.SetProgress(j.Id, ProgressNew)
				if err != nil {
					results <- result{fileId: j.Id, err: err}
				}
			}
			// wait for all workers to finish
			pwg.Wait()
			// read all the results
			close(results)
			for r := range results {
				log.Println("file", r.fileId, "err", r.err)
			}
			return
		}
	}
}

func (b *backgroundService) processFileWorker(id int, files <-chan File, results chan<- result, timeToStop <-chan bool, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		log.Println("worker", id, "stopped")
	}()

	var err error
	for {
		select {
		case f, ok := <-files:
			if !ok {
				return
			}
			log.Println("worker", id, "started  job", f.Id)
			err = b.ProcessFile(f)
			log.Println("worker", id, "finished job", f.Id)
			results <- result{fileId: f.Id, err: err}
		case <-timeToStop:
			return
		}
	}
}

func (b *backgroundService) ProcessFile(file File) error {

	err := b.repo.SetProgress(file.Id, ProgressInProgress)
	if err != nil {
		return err
	}

	// setting new file status
	defer func() {
		if err == nil {
			// setting error from nil to this
			err = b.repo.SetProgress(file.Id, ProgressCompleted)
		} else {
			_ = b.repo.SetProgress(file.Id, ProgressErrorEncountered)
		}
	}()

	f, err := b.storage.OpenFileUpFolder(file.InternalName)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()

	output, err := b.storage.CreateFileReadyFolder(file.InternalName + ".wav")
	if err != nil {
		return err
	}
	// todo go-lame
	err = wav.Encode(output, smoothing(smoothing(streamer)), format)
	if err != nil {
		return err
	}
	defer output.Close()

	err = b.storage.RemoveFileUpFolder(file.InternalName)
	if err != nil {
		return err
	}

	// todo https://github.com/viert/go-lame/blob/master/encoder_test.go encoding settings

	of, err := b.storage.CreateFileReadyFolder(file.InternalName)
	if err != nil {
		return err
	}
	defer of.Close()
	enc := lame.NewEncoder(of)
	defer enc.Close()

	inf, err := b.storage.OpenFileReadyFolder(file.InternalName + ".wav")
	if err != nil {
		return err
	}
	defer inf.Close()

	r := bufio.NewReader(inf)
	r.WriteTo(enc)

	err = b.storage.RemoveFileReadyFolder(file.InternalName + ".wav")
	if err != nil {
		return err
	}

	return nil
}

func smoothing(s beep.Streamer) beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {

		n, ok = s.Stream(samples)

		l := len(samples)

		for i := range samples {
			if i >= 1 && i < l-1 {
				samples[i][0] = (samples[i-1][0] + samples[i][0] + samples[i+1][0]) / 3
				samples[i][1] = (samples[i-1][1] + samples[i][1] + samples[i+1][1]) / 3
			}
		}

		return n, ok
	})
}

func addNoise(s beep.Streamer) beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {

		n, ok = s.Stream(samples)

		for i := range samples {
			samples[i][0] = samples[i][0] + rand.Float64()*0.025 - 0.01225
			samples[i][1] = samples[i][1] + rand.Float64()*0.025 - 0.01225
		}

		return n, ok
	})
}
