package main

import (
	handlers "r-docs/handler"
	"r-docs/ws"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	hub := ws.NewHub()
	go hub.Run()

	e.GET("/ws", func(c echo.Context) error {
		return handlers.HandleWebSocket(c, hub)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
