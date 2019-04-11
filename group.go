package executor

import "sync"

type IGroupMgr interface {
	Add(groupID int, key string, value interface{})
	Delete(groupID int)
}

type groupJob struct {
	groupID int
	key     string
	value   interface{}
}

type ExecutorGroup struct {
	*sync.RWMutex
	exec    *Executor
	groups  map[int]*Group
	counter int
}

func NewExecutorGroup(maxWorkers uint) *ExecutorGroup {
	me := &ExecutorGroup{
		RWMutex: &sync.RWMutex{},
		groups:  make(map[int]*Group),
	}

	exec := New(maxWorkers, 10, func(key string, value interface{}) {
		job := value.(groupJob)
		me.RLock()
		group := me.groups[job.groupID]
		me.RUnlock()
		group.Handle(key, job.value)
	})

	me.exec = exec
	return me
}

func (me *ExecutorGroup) Add(groupID int, key string, value interface{}) {
	me.exec.Add(key, groupJob{groupID: groupID, key: key, value: value})
}

func (me *ExecutorGroup) Delete(groupID int) {
	me.Lock()
	delete(me.groups, groupID)
	me.Unlock()
}

func (me *ExecutorGroup) NewGroup(handler func(string, interface{})) *Group {
	me.Lock()
	me.counter++
	id := me.counter
	group := &Group{id: id, mgr: me, barrier: &sync.Mutex{}, handler: handler, Mutex: &sync.Mutex{}}
	me.groups[id] = group
	me.Unlock()
	return group
}

type Group struct {
	*sync.Mutex
	id               int
	mgr              IGroupMgr
	barrier          *sync.Mutex
	numProcessingJob int
	handler          func(string, interface{})
}

func (me *Group) Add(key string, value interface{}) {
	me.Lock()
	me.numProcessingJob++
	if me.numProcessingJob == 1 {
		me.barrier.Lock()
	}
	me.Unlock()
	me.mgr.Add(me.id, key, value)
}

func (me *Group) Delete() {
	me.mgr.Delete(me.id)
}

func (me *Group) Handle(key string, value interface{}) {
	me.handler(key, value)

	me.Lock()
	me.numProcessingJob--
	if me.numProcessingJob == 0 {
		me.barrier.Unlock()
	}
	me.Unlock()
}

func (me *Group) Wait() {
	me.barrier.Lock()
	me.barrier.Unlock()
}
