package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	st := _main()
	os.Exit(st)
}

func _main() int {
	quit := make(chan struct{})

	wg := &sync.WaitGroup{}
	for _, pat := range os.Args[1:] {
		r, err := parseRule(pat)
		if err != nil {
			return 1
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
	exitStatus := 0
	for loop {
		select {
		case <-sigCh:
			close(quit)
			loop = false
			exitStatus = 1
		case <-done:
			loop = false
		}
	}

	return exitStatus
}