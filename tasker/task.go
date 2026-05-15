package tasker

import (
	"NUMParser/utils"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Tasker struct {
	tasks      []worker
	threads    int
	shuffle    bool
	disableLog bool
	isStop     atomic.Bool
	wa         sync.WaitGroup
	mu         sync.Mutex
}

type worker struct {
	Func func(interface{}) bool
	Data interface{}
}

func New(threads int, shuffle bool) *Tasker {
	t := new(Tasker)
	t.threads = threads
	if t.threads < 1 {
		t.threads = 1
	}
	t.shuffle = shuffle
	return t
}

func (t *Tasker) DisableLog() {
	t.disableLog = true
}

func (t *Tasker) Add(fn func(interface{}) bool, data interface{}) {
	t.mu.Lock()
	t.tasks = append(t.tasks, worker{
		Func: fn,
		Data: data,
	})
	t.wa.Add(1)
	t.mu.Unlock()
}

func (t *Tasker) Run() {
	if t.shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(t.tasks), func(i, j int) { t.tasks[i], t.tasks[j] = t.tasks[j], t.tasks[i] })
	}
	utils.PForLim(t.tasks, t.threads, func(i int, wrk worker) bool {
		stopped := t.isStop.Load()
		if !t.disableLog && !stopped {
			log.Println("Task", i+1, "/", len(t.tasks))
		}
		if !stopped {
			if !wrk.Func(wrk.Data) {
				t.isStop.Store(true)
			}
		}
		t.wa.Done()
		return true
	})
}

func (t *Tasker) Wait() {
	if len(t.tasks) > 0 {
		t.wa.Wait()
	}
}
