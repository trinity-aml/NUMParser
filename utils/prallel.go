package utils

import (
	"sync"
	"sync/atomic"
)

func PFor[T any](arr []T, fn func(i int, el T)) {
	var wg sync.WaitGroup
	wg.Add(len(arr))
	for i := range arr {
		go func(i int) {
			defer wg.Done()
			fn(i, arr[i])
		}(i)
	}
	wg.Wait()
}

func PForLim[T any](arr []T, lim int, fn func(int, T) bool) {
	var wg sync.WaitGroup
	var isBreak atomic.Bool
	limits := make(chan struct{}, lim)
	for i := range arr {
		if isBreak.Load() {
			break
		}
		limits <- struct{}{}
		if isBreak.Load() {
			<-limits
			break
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() { <-limits }()
			if isBreak.Load() {
				return
			}
			if !fn(i, arr[i]) {
				isBreak.Store(true)
			}
		}(i)
	}
	wg.Wait()
}

func ParallelFor(begin, end int, fn func(i int)) {
	for i := begin; i < end; i++ {
		fn(i)
	}
}

func ParallelLimFor(begin, end, lim int, fn func(i int)) {
	var wg sync.WaitGroup
	wg.Add(end - begin)
	limits := make(chan struct{}, lim)
	for i := begin; i < end; i++ {
		limits <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-limits }()
			fn(i)
		}(i)
	}
	wg.Wait()
}
