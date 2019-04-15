package executor

import "sync"

type groupJob struct {
	groupID int
	key     string
	value   interface{}
}
type GroupMgr struct {
	*sync.Mutex
	exec   *Executor
	groups map[int]*Group

	// used to alloc group id, increases each time a new group is added
	counter int
}

// NewGroupMgr creates a new GroupMgr object
func NewGroupMgr(maxWorkers uint) *GroupMgr {
	me := &GroupMgr{Mutex: &sync.Mutex{}, groups: make(map[int]*Group)}

	me.exec = New(maxWorkers, func(key string, value interface{}) {
		job := value.(groupJob)
		me.Lock()
		group := me.groups[job.groupID]
		me.Unlock()
		group.handle(key, job.value)
	})
	return me
}

// add is called by a group to registers a job
func (me *GroupMgr) addJob(groupID int, key string, value interface{}) {
	me.exec.Add(key, groupJob{groupID: groupID, key: key, value: value})
}

// deleteGroup is called by a group to release itself from the manager
func (me *GroupMgr) deleteGroup(groupID int) {
	me.Lock()
	delete(me.groups, groupID)
	me.Unlock()
}

// NewGroup registers a group to the manager
func (me *GroupMgr) NewGroup(handler func(string, interface{})) *Group {
	me.Lock()
	me.counter++
	id := me.counter
	group := &Group{lock: me.Mutex, id: id, mgr: me, barrier: &sync.Mutex{}, handler: handler}
	me.groups[id] = group
	me.Unlock()
	return group
}

type Group struct {
	id               int
	mgr              *GroupMgr
	barrier          *sync.Mutex

	lock             *sync.Mutex // protect numProcessingJob
	numProcessingJob int // holds current number of processing jobs

	handler          func(string, interface{})
}

// Add adds a new job.
// Caution: calling a released (deleted) group have no effect. User should make
// sure that Add is alway called before Wait
func (me *Group) Add(key string, value interface{}) {
	me.lock.Lock()
	me.numProcessingJob++
	if me.numProcessingJob == 1 {
		me.barrier.Lock()
	}
	me.lock.Unlock()
	me.mgr.addJob(me.id, key, value)
}

// handler is used by the manager to call users' handler function
func (me *Group) handle(key string, value interface{}) {
	me.handler(key, value)

	me.lock.Lock()
	me.numProcessingJob--
	if me.numProcessingJob == 0 {
		me.barrier.Unlock()
	}
	me.lock.Unlock()
}

// Wait blocks current caller until there is no processing jobs. This function
// is must be call after you done with the group. Otherwise the group stay forever
// Note: This function also release the current group, future calls to Add
// will be ignore.
func (me *Group) Wait() {
	me.barrier.Lock()
	me.barrier.Unlock()
	me.mgr.deleteGroup(me.id)
}
