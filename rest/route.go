package rest

import (
	"github.com/labstack/echo"
)

func Route(e *echo.Echo) {
	e.GET("/hello", getReqHello)
	e.GET("/auth/tempsession", getReqTempSession)
	e.GET("/auth/session", getReqSession)
	e.DELETE("/auth/session", delReqSession)
	e.POST("/user", postReqUser)
	e.GET("/user/:id", getReqUserData)
}
