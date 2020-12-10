package workflowrun

import (
	"container/heap"
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/common"
)

var (
	maxRetry       = 10
	defaultTimeout = time.Hour * 2
	timeStep       = time.Second * 10
	errClosed      = errors.New("blockingStage queue close")
	errNotFound    = errors.New("stage not found")
	errRemoved     = errors.New("blockingStage was removed")
	stageChan      chan *BlockingStage
)

func init() {
	stageChan = make(chan *BlockingStage)
}

// BlockingStage represent s blocked stage information
type BlockingStage struct {
	workflow    *v1alpha1.Workflow
	workflowRun *v1alpha1.WorkflowRun
	stage       *v1alpha1.Stage
	blockTime   time.Time
	expireTime  time.Time
	retry       int
}

// NewBlockingStage create a blocking
func NewBlockingStage(wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun, stg *v1alpha1.Stage) *BlockingStage {
	timeout, err := ParseTime(wfr.Spec.Timeout)
	if err != nil {
		timeout = defaultTimeout
	}
	now := time.Now()
	return &BlockingStage{
		workflow:    wf,
		workflowRun: wfr,
		stage:       stg,
		blockTime:   now,
		expireTime:  now.Add(timeout),
	}
}

type heapItem struct {
	obj   *BlockingStage
	index int
}

type heapData struct {
	items map[string]*heapItem
	queue []string
}

var (
	_ = heap.Interface(&heapData{})
)

// Less ...
func (h *heapData) Less(i, j int) bool {
	if i > len(h.queue) || j > len(h.queue) {
		return false
	}
	itemi, ok := h.items[h.queue[i]]
	if !ok {
		return false
	}
	itemj, ok := h.items[h.queue[j]]
	if !ok {
		return false
	}

	return itemi.obj.blockTime.Before(itemj.obj.blockTime)
}

// Len ...
func (h *heapData) Len() int { return len(h.queue) }

// Swap ...
func (h *heapData) Swap(i, j int) {
	h.queue[i], h.queue[j] = h.queue[j], h.queue[i]
	item := h.items[h.queue[i]]
	item.index = i
	item = h.items[h.queue[j]]
	item.index = j
}

// Push ...
func (h *heapData) Push(elem interface{}) {
	blkStg := elem.(*BlockingStage)
	n := len(h.queue)
	h.items[blkStg.stage.Name] = &heapItem{
		obj:   blkStg,
		index: n,
	}
	h.queue = append(h.queue, blkStg.stage.Name)
}

// Pop ...
func (h *heapData) Pop() interface{} {
	key := h.queue[len(h.queue)-1]
	h.queue = h.queue[0 : len(h.queue)-1]
	item, ok := h.items[key]
	if !ok {
		return nil
	}
	delete(h.items, key)
	return item.obj
}

// PriorityQueue priority queue
type PriorityQueue struct {
	lock   sync.RWMutex
	cond   sync.Cond
	data   *heapData
	closed bool
}

// Close close queue
func (h *PriorityQueue) Close() {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.closed = true
	h.cond.Broadcast()
}

// Add add blocking stage
func (h *PriorityQueue) Add(stg *BlockingStage) error {
	key := stg.stage.Name
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.closed {
		return errClosed
	}
	if _, exists := h.data.items[key]; exists {
		h.data.items[key].obj = stg
		heap.Fix(h.data, h.data.items[key].index)
	} else {
		h.addIfNotPresentLocked(key, stg)
	}
	h.cond.Broadcast()
	return nil
}

// BulkAdd add blocking stage
func (h *PriorityQueue) BulkAdd(list []*BlockingStage) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.closed {
		return errClosed
	}
	for _, obj := range list {
		key := obj.stage.Name
		if _, exists := h.data.items[key]; exists {
			h.data.items[key].obj = obj
			heap.Fix(h.data, h.data.items[key].index)
		} else {
			h.addIfNotPresentLocked(key, obj)
		}
	}
	h.cond.Broadcast()
	return nil
}

// AddIfNotPresent if not present add
func (h *PriorityQueue) AddIfNotPresent(obj *BlockingStage) error {
	key := obj.stage.Name
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.closed {
		return errClosed
	}
	h.addIfNotPresentLocked(key, obj)
	h.cond.Broadcast()
	return nil
}

func (h *PriorityQueue) addIfNotPresentLocked(key string, obj *BlockingStage) {
	if _, exists := h.data.items[key]; exists {
		return
	}
	heap.Push(h.data, obj)
}

// Update update queue by add blocking stage
func (h *PriorityQueue) Update(obj *BlockingStage) error {
	return h.Add(obj)
}

// Delete delete blocking stage from queue
func (h *PriorityQueue) Delete(obj *BlockingStage) {
	key := obj.stage.Name
	h.lock.Lock()
	defer h.lock.Unlock()
	if item, ok := h.data.items[key]; ok {
		heap.Remove(h.data, item.index)
	}
}

// Pop pop blocking stage from queue
func (h *PriorityQueue) Pop() (*BlockingStage, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	for len(h.data.queue) == 0 {
		if h.closed {
			return nil, errClosed
		}
		h.cond.Wait()
	}
	obj := heap.Pop(h.data)
	if obj != nil {
		return obj.(*BlockingStage), nil
	}

	return nil, errRemoved
}

// List list blocking stage in queue
func (h *PriorityQueue) List() []*BlockingStage {
	h.lock.RLock()
	defer h.lock.RUnlock()
	list := make([]*BlockingStage, 0, len(h.data.items))
	for _, item := range h.data.items {
		list = append(list, item.obj)
	}
	return list
}

// ListKeys list keys in queue
func (h *PriorityQueue) ListKeys() []string {
	h.lock.RLock()
	defer h.lock.RLock()
	list := make([]string, 0, len(h.data.items))
	for key := range h.data.items {
		list = append(list, key)
	}
	return list
}

// GetByKey get blocking stage by key from queue
func (h *PriorityQueue) GetByKey(key string) (*BlockingStage, error) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	item, exists := h.data.items[key]
	if !exists {
		return nil, errNotFound
	}
	return item.obj, nil
}

// IsClosed ...
func (h *PriorityQueue) IsClosed() bool {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.closed
}

// NewPriorityQueue ...
func NewPriorityQueue() *PriorityQueue {
	h := &PriorityQueue{
		data: &heapData{
			items: map[string]*heapItem{},
			queue: []string{},
		},
	}
	h.cond.L = &h.lock
	return h
}

// BlockingStageProcessor processor to process blocking stage
type BlockingStageProcessor struct {
	client             clientset.Interface
	stagePriorityQueue *PriorityQueue
	stopChan           chan struct{}
	sync.WaitGroup
}

// NewBlockingStageProcessor ...
func NewBlockingStageProcessor(client clientset.Interface) *BlockingStageProcessor {
	pq := NewPriorityQueue()
	processor := &BlockingStageProcessor{
		client:             client,
		stagePriorityQueue: pq,
		stopChan:           make(chan struct{}),
	}
	go processor.Run()

	return processor
}

// Run ...
func (h *BlockingStageProcessor) Run() {
	go func() {
		for {
			select {
			case stg := <-stageChan:
				log.WithField("blocking stage", stg.stage.Name).
					WithField("wf", stg.workflow.Name).
					WithField("wfr", stg.workflowRun.Name).
					Info("add to the queue")
				_ = h.stagePriorityQueue.Add(stg) // TODO error process
			case <-h.stopChan:
				return
			}
		}
	}()

	for {
		blkStg, err := h.stagePriorityQueue.Pop()
		switch err {
		case errClosed:
			return
		case errRemoved:
			continue
		}
		if blkStg.expireTime.Before(time.Now()) {
			// just remove it from queue
			// because timeout processor can do clean job in a proper time
			log.WithField("blocking stage", blkStg.stage.Name).
				WithField("expire time", blkStg.expireTime).
				Warn("timeout")
			continue
		}
		h.Add(1)
		h.process(blkStg)
	}
}

// Stop ...
func (h *BlockingStageProcessor) Stop() {
	h.Wait()
	close(h.stopChan)
	h.stagePriorityQueue.Close()
}

func (h *BlockingStageProcessor) process(stg *BlockingStage) {
	defer h.Done()
	clusterClient := common.GetExecutionClusterClient(stg.workflowRun)
	operator, err := NewOperator(clusterClient, h.client, stg.workflowRun, stg.workflowRun.Namespace)
	if err != nil {
		log.WithField("blocking stage", stg.stage.Name).
			Errorf("failed to create workflowRun operator %v\n", err)
		h.stagePriorityQueue.Delete(stg)
	}
	err = NewWorkloadProcessor(clusterClient, h.client, stg.workflow, stg.workflowRun, stg.stage, operator).
		Process()
	if isExceedResourceQuotaError(err) {
		// the first blocked stage in priorityQueue can try to run maxRetry times,
		// if it still failed because of quota exceeded
		// we reduce the priority by adding a time step and reset it's retry count
		// so that the other blocked stage behind it which required less resource can get the chance to run
		stg.retry++
		if stg.retry >= maxRetry {
			stg.retry = 0
			stg.blockTime.Add(timeStep)
		}
		if stg.blockTime.After(time.Now()) {
			stg.blockTime = time.Now()
		}
		_ = h.stagePriorityQueue.Update(stg) // TODO error process
		err = nil
	}
	if err != nil {
		log.WithField("blocking stage", stg.stage.Name).
			Errorf("workload processor failed process %v\n", err)
		h.stagePriorityQueue.Delete(stg)
	}
	// update wfr final status
	err = operator.Update()
	if err != nil {
		log.WithField("blocking stage", stg.stage.Name).
			Errorf("operator update status failed %v\n", err)
	}
}
