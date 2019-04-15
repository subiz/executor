package executor

import "sync"

type IGroupMgr interface {
	Add(groupID int, key string, value interface{})
	Delete(groupID int)
	Lock()
	Unlock()
}

type groupJob struct {
	groupID int
	key     string
	value   interface{}
}

type ExecutorGroup struct {
	*sync.Mutex
	exec      *Executor
	groups    map[int]*Group
	counter   int
	groupLock *sync.Mutex
}

func NewExecutorGroup(maxWorkers uint) *ExecutorGroup {
	me := &ExecutorGroup{
		Mutex:   &sync.Mutex{},
		groups:    make(map[int]*Group),
		groupLock: &sync.Mutex{},
	}

	exec := New(maxWorkers, 10, func(key string, value interface{}) {
		job := value.(groupJob)
		me.Lock()
		group := me.groups[job.groupID]
		me.Unlock()
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
	group := &Group{id: id, mgr: me, barrier: &sync.Mutex{}, handler: handler}
	me.groups[id] = group
	me.Unlock()
	return group
}

func (me *ExecutorGroup) Lock() {
	me.groupLock.Lock()
}

func (me *ExecutorGroup) Unlock() {
	me.groupLock.Unlock()
}

type Group struct {
	id               int
	mgr              IGroupMgr
	barrier          *sync.Mutex
	numProcessingJob int
	handler          func(string, interface{})
}

// Add adds a new job.
// Caution: calling a released (deleted) group have no effect. User should make
// sure that Add is alway called before Wait
func (me Group) Add(key string, value interface{}) {
	me.mgr.Lock()
	me.numProcessingJob++
	if me.numProcessingJob == 1 {
		me.barrier.Lock()
	}
	me.mgr.Unlock()
	me.mgr.Add(me.id, key, value)
}

// Handler is used by the manager to call users' handler function
func (me *Group) Handle(key string, value interface{}) {
	me.handler(key, value)

	me.mgr.Lock()
	me.numProcessingJob--
	if me.numProcessingJob == 0 {
		me.barrier.Unlock()
	}
	me.mgr.Unlock()
}

// Wait blocks current caller until there is no processing jobs.
// Note: This function also release the current group, future calls to Add
// will be ignore. So make sure this is the last function you call after done
// with the group
func (me *Group) Wait() {
	me.barrier.Lock()
	me.barrier.Unlock()
	me.mgr.Delete(me.id)
}
