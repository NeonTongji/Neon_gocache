package lru

import "container/list"

// Cache is a LRU cache. It is not safe for concurrent access.
/**
cache 结构体，包含了各种成员变量信息，类似于java 成员变量域的集合
*/
type Cache struct {
	maxBytes int64                    // cache的最大容量
	nbytes   int64                    // 当前容量
	ll       *list.List               // cache关联的一个双向链表 list初始化可以list.New(), 或者list.List
	cache    map[string]*list.Element // *list.Element是list中的一个节点，go中的list的每个元素get返回类型都是一个list.Element类型
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

/**
一个kv对定义为一个entry
*/
type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
// 用了记录占了多少字节，这里是一个接口，内部是Len()函数
type Value interface {
	Len() int
}

// New is the Constructor of Cache
// New是cache的构造器
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get look ups a key's value
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
