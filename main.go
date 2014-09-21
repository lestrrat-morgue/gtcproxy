package main

import (
	"fmt"
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
		p, err := parseRule(pat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bad rule: %s", err)
			return 1
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p.Run(quit); err != nil {
				fmt.Printf("Failed to run proxy: %s\n", err)
			}
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