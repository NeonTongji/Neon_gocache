package neoncache

import (
	"fmt"
	"log"
	"sync"
)

// Group作为分离变量的命名空间
type Group struct {
	name      string //命名空间名称
	getter    Getter // 缓存未命中时，获取源数据的回调
	mainCache cache  // 并发缓存
}

// Getter封装了Get方法
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get方法 实现了Getter接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex              //读写锁
	groups = make(map[string]*Group) // groups为一个map，key是string类型，value为Group
)

// 创新一个新的Group实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	// 加锁
	mu.Lock()
	defer mu.Unlock()
	// 根据Group的构造器，初始化g
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	// groups中新增 kv对 name, g
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.Unlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("缺少key")
	}

	// (1) 从maincache中获取，
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Neon cache] hit")
		return v, nil
	}
	// (2) 缓存不存在，则调用 load 方法，
	// load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取），
	// getLocally 调用用户回调函数 g.getter.Get() 获取源数据，
	// 并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
	return g.load(key)
}

// load从本地缓存获取
func (g *Group) load(key string) (ByteView, error) {
	// load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取），
	return g.getLocally(key)
}

// getLocally 调用用户回调函数 g.getter.Get() 获取源数据，
// 并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	//value为bytes的副本
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value) // 存到maincache中
	return value, nil
}

// 将kv新增到maincache中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
