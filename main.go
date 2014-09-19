package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	quit := make(chan struct{})

	wg := &sync.WaitGroup{}
	for _, pat := range os.Args[1:] {
		r, err := parseRule(pat)
		if err != nil {
			return
		}

		p := NewProxy(r.src, r.dst)
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Run(quit)
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	loop := true
	for loop {
		select {
		case <-sigCh:
			close(quit)
			loop = false
		case <-done:
			loop = false
		}
	}

}