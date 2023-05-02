package main

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const DefTime = "2006-01-02 15:04:05"

type cip struct {
	Ip   string
	Time string
}

type sip struct {
	sync.RWMutex
	User map[string]cip
}

func main() {
	app := fiber.New()
	users := sip{User: make(map[string]cip)}

	//查看所有用户的IP信息
	app.Get("/list", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"data": users.User})
	})

	//提交用户的IP
	app.Get("/dns/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		if name != "" {
			users.Lock()
			users.User[name] = cip{Ip: c.IP(), Time: time.Now().Format(DefTime)}
			users.Unlock()
		}
		return c.SendString("OK")
	})

	//监听端口2178
	go app.Listen(":2178") //这里不在协程中运行，错误不会出现

	for {
		time.Sleep(time.Second)
	}
}
