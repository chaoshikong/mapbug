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
就是说，ctx.Params()返回的是个引用，需要使用复制功能，难怪fasthttp比官方库net/http速度更快，连string都使用的是引用返回的么，一来一回几次对话老外有些急了，直接给出解决方案了，
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
就是要使用utils.CopyString()把引用再复制1份，要不然有冲突，不过深层次的原因，就是要去读fiber的源代码了
