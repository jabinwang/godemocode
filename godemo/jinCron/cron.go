package jinCron

import (
	"sort"
	"time"
)

type Schedule interface {
	next(t time.Time) time.Time
}

func Every(d time.Duration) Schedule{
	return & sche{
		d: d,
	}
}

type sche struct{
	d time.Duration
}

func (s *sche) next(t time.Time) time.Time {
	return t.Truncate(time.Second).Add(s.d)
}

type Entry struct {
	Schedule Schedule
	Job Job
	Next time.Time
	Prev time.Time
}

type Job interface {
	Run()
}

type JinCron struct {
	entries []*Entry
	running bool
	add chan *Entry
	stop chan struct{}
}

func NewCron() *JinCron {
	return &JinCron{
		add: make(chan *Entry),
		stop: make(chan struct{}),
	}
}

func (c *JinCron) Add(s Schedule, j Job)  {
	entry := &Entry{
		Schedule: s,
		Job: j,
	}

	if !c.running {
		c.entries = append(c.entries, entry)
		return
	}
	c.add <-entry
}

func (c *JinCron) AddFunc(s Schedule, j func())  {
	c.Add(s, JobFunc(j))
}

func (c *JinCron) Start()  {
	c.running = true
	go c.run()
}

func (c *JinCron) Stop()  {
	if !c.running {
		return
	}
	c.running = false
	c.stop <- struct {}{}
}

var after = time.After
func (c *JinCron) run() {
	var t time.Time
	now := time.Now().Local()
	for _, e := range c.entries {
		e.Next = e.Schedule.next(now)
	}
	for  {
		sort.Sort(byTime(c.entries))
		if len(c.entries) > 0 {
			t = c.entries[0].Next
		} else {

		}
		select {
		case now = <-after(t.Sub(now)):
			for _, entry := range c.entries {
				if entry.Next != t {
					break
				}
				entry.Prev = now
				entry.Next = entry.Schedule.next(now)
				go entry.Job.Run()
			}

		case e := <-c.add:
			e.Next = e.Schedule.next(time.Now())
			c.entries = append(c.entries, e)

			case <- c.stop:
				return
		}
	}
}

type byTime []*Entry

func (b byTime) Len() int {
	return len(b)
}

func (b byTime) Less(i, j int) bool {
	if b[i].Next.IsZero() {
		return false
	}
	if b[j].Next.IsZero() {
		return true
	}
	return b[i].Next.Before(b[j].Next)
}

func (b byTime) Swap(i, j int) {
	b[i],b[j]= b[j],b[i]
}

type JobFunc func()

func (j JobFunc) Run()  {
	j()
}