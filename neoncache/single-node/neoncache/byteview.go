package neoncache

// ByteView是一个不可变byte数组
type ByteView struct {
	b []byte // 选用byte便可以支持任何数据类型
}

// 返回ByteView数组v的长度
func (v ByteView) Len() int {
	return len(v.b)
}

// 返回ByteView数组的一个拷贝，b才是真是数据，防止b被外部篡改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// String讲ByteView数组作为字符串返回，必要时会生成一份副本
func (v ByteView) String() string {
	return string(v.b)
}
