package backup

import (
	"context"
	"sync"
)

type Pool struct {
	Size       int
	RetryCount int
	ch         chan Work
}

func NewPool(size int) *Pool {
	return &Pool{
		Size: size,
		ch:   make(chan Work, size),
	}
}

func (p *Pool) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	cherr := make(chan error, p.Size)
	ctx, cancel := context.WithCancel(ctx)

	// create p.Size workers as goroutines
	wg.Add(p.Size)
	for i := 0; i < p.Size; i++ {
		go func() {
			if err := p.startLoop(ctx); err != nil {
				cherr <- err
			}
		}()
	}

	// I also want to remember the *first* error and cancel the context accordingly
	var err error
	go func() {
		err, _ = <-cherr
		cancel()
	}()

	// and wait for completion
	wg.Wait()
	close(cherr)
	return err
}

func (p *Pool) EnqueueDowload(ctx context.Context, cfg *Config, url, filePath string) error {
	payload := &DownloadPayload{
		URL:      url,
		FilePath: filePath}
	return p.enqueue(ctx, Work{WorkTypeDownload, cfg, payload})
}

func (p *Pool) EnqueueWrite(ctx context.Context, cfg *Config, filePath string, contents []byte) error {
	payload := &WriteFilePayload{
		FilePath: filePath,
		Contents: contents,
	}
	return p.enqueue(ctx, Work{WorkTypeWriteFile, cfg, payload})
}

func (p *Pool) enqueue(ctx context.Context, work Work) error {
	select {
	case p.ch <- work:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (p *Pool) Done() {
	close(p.ch)
}

// startLoop loops until either the context is canceled or the input channel is closed.
func (p *Pool) startLoop(ctx context.Context) error {
	var err error
Loop:
	for {
		select {
		case work, ok := <-p.ch:
			if !ok {
				break Loop // input channel closed !
			}
		RetryLoop:
			for i := 0; i < p.RetryCount+1; i++ {
				if err = work.Run(ctx); err == nil {
					break RetryLoop // no error => keep going
				}
			}
			if err != nil {
				break Loop // we run once + p.RetryCount retries and still got and eror !
			}
		case <-ctx.Done():
			err = ctx.Err() // Canceled or DeadlineExceeded
			break Loop
		}
	}
	return err
}
