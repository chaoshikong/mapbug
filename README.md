# mapbug
登记一个go语言中使用map的bug

本项目是在写gofiber项目的时候发现的一个BUG，就是会登记客户端上传的IP地址及当时的时间的功能时发现的map中存在问题

运行本项目后，浏览器多次输入http://127.0.0.1:2178/dns/strings，strings表示不同长度的字符串，多输几次后

再到浏览器中用http://127.0.0.1:2178/list 查看的时候，发现提交的多次数据中，有的数据错掉了

比如我提交了"abcde"，然后又提交了"efg"，用list查询的时候，发现原来提交的"abcde"变成"efgde"了

已提交了这个BUG，地址在https://github.com/golang/go/issues/59917 等待回复中
几个小时后，得到的回复是：
I believe this is from fiber being unsafe rather an issue with go itself.我认为这是因为fiber不安全，而不是go本身的问题。

于是我把fiber框架换成gin，结果真的就没有问题了。。。看来还真是fiber的问题

那么，去fiber网去提问题，地址https://github.com/gofiber/fiber/issues/2446 等待回复中

然后回复的The values given by the ctx.Params method are mutable (also a reference)，Pls use the copy function before you store it
就是说，ctx.Params()返回的是个引用，需要使用复制功能，难怪fiber比使用官方库net/http的gin速度更快，连string都使用的是引用返回的么，一来一回几次对话老外有些急了，直接给出解决方案了，
代码如下：
```go
	app.Get("/dns/:name", func(c *fiber.Ctx) error {
		name := utils.CopyString(c.Params("name"))
		if name != "" {
			users.Lock()
			users.User[name] = cip{Ip: c.IP(), Time: time.Now().Format(DefTime)}
			users.Unlock()
		}
		return c.SendString("OK")
	})
```
就是要使用utils.CopyString()把引用再复制1份，要不然有冲突，然后我去查了一下gin的Param()函数，和fiber.Params()有什么不同时，发现
gin使用的是func (ps Params) ByName(name string) (va string) {}，因为不是(ps * Params)也就是说把参数复制了一份
而fiber.Params()中用的是func (c * Ctx) Params(key string, defaultValue ...string) string {}，不知道和这个有没有关系
再深入点了解，发现了一段这样的代码，很能说明问题：
```go
package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	//app := fiber.New(fiber.Config{Immutable: true})//如果启用Immutable:true，则没有任何问题，否则3秒后的值是会变的
	app := fiber.New()

	app.Get("/:number", func(c *fiber.Ctx) error {
		number := c.Params("number")
		go myfunc(number)
		return c.SendString(number)
	})
	app.Listen(":3000")
}

func myfunc(number string) {
	fmt.Printf("number is %s \n", number)
	time.Sleep(3 * time.Second)
	fmt.Printf("number is now %s \n", number)
}
```
看来fiber为了做到速度极致，把Immutable的值默认为false，因为true时性能会下降，这样就可以根据用户的需求来选择后期是否要复制，或者前期取true，两种选择
不过了解得越多，我也越来越喜欢fiber了，因为做得真的很人性化，对于我这种喜欢追求速度的人来说。。。

然后我又深挖，看看utils.CopyString中的代码
```go
type StringHeader struct {
	Data uintptr
	Len  int
}

func UnsafeBytes(s string) []byte {
	if s == "" {
		return nil
	}

	return (*[MaxStringLen]byte)(unsafe.Pointer(
		(*reflect.StringHeader)(unsafe.Pointer(&s)).Data),
	)[:len(s):len(s)]
}

func CopyString(s string) string {
	return string(UnsafeBytes(s))
}
```
可以看到，在GO中的string类型，其实是一个struct{Data uintptr,Len int}结构，那么完全有可能1个字符串赋值给另一个字符串的时候，不用真正的复制，而是只要改变内部的指针即可，写个代码来证明我的猜想试试
```go
func main() {
	var a, b string
	a = "abc"
	prints(a)
	b = a
	prints(b)
	a = "aaa"
	prints(a)
	b = "efg"
	prints(b)
}

func prints(s string) {
	fmt.Println((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)
}
结果：
8801960
8801960
8801957
8801972
```
看来是go在的string的问题，真象大白了，原来go为了保证运行速度，string赋值时并不是复制内容，而是指向同一块内存，这就解释了map[name]时，会把原来的值改掉的原因了。
