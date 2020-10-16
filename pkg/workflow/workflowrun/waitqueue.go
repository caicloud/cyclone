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
	blockingChan   chan *BlockingStage
)

func init() {
	blockingChan = make(chan *BlockingStage)
}

type BlockingStage struct {
	workflow    *v1alpha1.Workflow
	workflowRun *v1alpha1.WorkflowRun
	stage       *v1alpha1.Stage
	blockTime   time.Time
	expireTime  time.Time
	retry       int
}

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

func (h *heapData) Len() int { return len(h.queue) }

func (h *heapData) Swap(i, j int) {
	h.queue[i], h.queue[j] = h.queue[j], h.queue[i]
	item := h.items[h.queue[i]]
	item.index = i
	item = h.items[h.queue[j]]
	item.index = j
}

func (h *heapData) Push(elem interface{}) {
	blkStg := elem.(*BlockingStage)
	n := len(h.queue)
	h.items[blkStg.stage.Name] = &heapItem{
		obj:   blkStg,
		index: n,
	}
	h.queue = append(h.queue, blkStg.stage.Name)
}

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

type PriorityQueue struct {
	lock   sync.RWMutex
	cond   sync.Cond
	data   *heapData
	closed bool
}

func (h *PriorityQueue) Close() {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.closed = true
	h.cond.Broadcast()
}

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

func (h *PriorityQueue) Update(obj *BlockingStage) error {
	return h.Add(obj)
}

func (h *PriorityQueue) Delete(obj *BlockingStage) {
	key := obj.stage.Name
	h.lock.Lock()
	defer h.lock.Unlock()
	if item, ok := h.data.items[key]; ok {
		heap.Remove(h.data, item.index)
	}
}

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
	} else {
		return nil, errRemoved
	}
}

func (h *PriorityQueue) List() []*BlockingStage {
	h.lock.RLock()
	defer h.lock.RUnlock()
	list := make([]*BlockingStage, 0, len(h.data.items))
	for _, item := range h.data.items {
		list = append(list, item.obj)
	}
	return list
}

func (h *PriorityQueue) ListKeys() []string {
	h.lock.RLock()
	defer h.lock.RLock()
	list := make([]string, 0, len(h.data.items))
	for key := range h.data.items {
		list = append(list, key)
	}
	return list
}

func (h *PriorityQueue) GetByKey(key string) (*BlockingStage, error) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	item, exists := h.data.items[key]
	if !exists {
		return nil, errNotFound
	}
	return item.obj, nil
}

func (h *PriorityQueue) IsClosed() bool {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if h.closed {
		return true
	}
	return false
}

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

type BlockingStageProcessor struct {
	client             clientset.Interface
	stagePriorityQueue *PriorityQueue
	stopChan           chan struct{}
	sync.WaitGroup
}

// List all workflowRuns in InQueue Status from apiServer (in case of the process tear down)
// traverse the workflowRun's all stages, find out the Stage which lead to the workflowRun become InQueue status
// call AddStage add the InQueue status Stage to priority queue according it's creationTime
// when a workflowRun's Stage create pod failed due to resource exceeded quota, we also add the stage to the priority queue
// go start a background goroutine loop doing below works:
// query resource quotas in ns
// if the resources sufficient, get the front Stage from priority queue, create pod for it
// if success, remove the Stage in priorityQueue, update Stage and WorkflowRun's Status from InQueue status to Running Status
// else do nothing
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

func (h *BlockingStageProcessor) Run() {
	go func() {
		for {
			select {
			case stg := <-blockingChan:
				log.WithField("blocking stage", stg.stage.Name).
					WithField("wf", stg.workflow.Name).
					WithField("wfr", stg.workflowRun.Name).
					Info("add to the queue")
				h.stagePriorityQueue.Add(stg)
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
		/*
			go func() {
				defer h.Done()
				log.WithField("blocking stage", blkStg.stage.Name).
					WithField("wf", blkStg.workflow.Name).
					WithField("wfr", blkStg.workflowRun.Name).
					Info("start process")
				err := h.process(blkStg)
				log.WithField("blocking stage", blkStg.stage.Name).
					Errorf("process block stage failed %v\n", err)
			}()
		*/
		_ = h.process(blkStg)
	}
}

func (h *BlockingStageProcessor) Stop() {
	h.Wait()
	close(h.stopChan)
	h.stagePriorityQueue.Close()
}

func (h *BlockingStageProcessor) process(stg *BlockingStage) error {
	defer h.Done()
	clusterClient := common.GetExecutionClusterClient(stg.workflowRun)
	operator, err := NewOperator(clusterClient, h.client, stg.workflowRun, stg.workflowRun.Namespace)
	if err != nil {
		log.WithField("blocking stage", stg.stage.Name).
			Errorf("failed to create workflowRun operator %v\n", err)
		h.stagePriorityQueue.Delete(stg)
		return err
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
		_ = h.stagePriorityQueue.Update(stg) // TODO
		err = nil
	}
	if err != nil {
		log.WithField("blocking stage", stg.stage.Name).
			Errorf("workload processor failed process %v\n", err)
		h.stagePriorityQueue.Delete(stg)
		return err
	}
	err = operator.Update()
	if err != nil {
		log.WithField("blocking stage", stg.stage.Name).
			Errorf("operator update status failed %v\n", err)
	}

	return err
}
