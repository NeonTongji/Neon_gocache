package neoncache

import (
	"neon_gocache/neoncache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex // 互斥锁
	lru        *lru.Cache // 自己撸的lru缓存
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()         //先给缓存上个锁
	defer c.mu.Unlock() // defer常被用于打开、关闭、连接、断开连接、加锁、释放锁
	// 初始化一个lru
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}

	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
