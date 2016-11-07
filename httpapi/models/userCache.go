package models

import (
	"container/list"
	"sync"
)

type CacheMem struct {
	sync.RWMutex
	limits    int // 0 means unlimits
	cache     map[string]*list.Element
	lru       *list.List
}

var UserCache *CacheMem

func InitUserCache(limits int) {
	UserCache = NewCacheMem(limits)
}

type entry struct {
	key  string
	user *User
}

func NewCacheMem(limits int) *CacheMem {
	return &CacheMem{
		limits: limits,
		lru:    list.New(),
		cache:  make(map[string]*list.Element),
	}
}

func (c *CacheMem) Get(key string) (*User, bool) {
	c.RLock()
	defer c.RUnlock()

	if ue, ok := c.cache[key]; ok {

		if ue == nil {
			return nil, false
		}
		c.lru.MoveToFront(ue)

		e := ue.Value.(*entry)
		if e != nil && e.key == key {
			return e.user, true
		}
	}

	return nil, false
}

func (c *CacheMem) Add(key string, u *User) error {
	c.Lock()
	defer c.Unlock()
	var e *list.Element

	if _, ok := c.cache[key]; ok {
		return nil
	}

	e = c.lru.PushFront(&entry{
		key:  key,
		user: u,
	})

	c.cache[key] = e

	if c.limits != 0 && c.lru.Len() > c.limits {
		c.removeOldest()
	}
	return nil
}

func (c *CacheMem) Modify(key string, u *User) error {
	c.Lock()
	defer c.Unlock()
	var e *list.Element

	if e, ok := c.cache[key]; ok {
		e.Value.(*entry).user = u
		return nil
	}

	e = c.lru.PushFront(&entry{
		key:  key,
		user: u,
	})

	c.cache[key] = e

	if c.limits != 0 && c.lru.Len() > c.limits {
		c.removeOldest()
	}
	return nil
}

// Remove removes the provided key from the keywordCache.
func (c *CacheMem) Remove(key string) (err error) {
	c.Lock()
	defer c.Unlock()

	if e, hit := c.cache[key]; hit {
		c.removeElement(e)
	}
	return
}

func (c *CacheMem) RemoveAll() error {
	c.Lock()
	defer c.Unlock()

	for e := c.lru.Front(); e != nil; {
		e = c.removeElement(e)
	}

	return nil
}

// RemoveOldest removes the oldest item from the keywordCache.
func (c *CacheMem) RemoveOldest() {
	c.Lock()
	defer c.Unlock()
	c.removeOldest()
	return
}
func (c *CacheMem) removeOldest() {
	if c.cache == nil {
		return
	}
	e := c.lru.Back()
	if e != nil {
		c.removeElement(e)
	}
}

func (c *CacheMem) removeElement(e *list.Element) *list.Element {
	eNext := e.Next()
	c.lru.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	return eNext
}
