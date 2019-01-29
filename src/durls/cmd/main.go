package main

import (
	durls "durls/internal/durls"
	"fmt"
	"github.com/kataras/iris"
	_ "github.com/kataras/iris/context"
	"log"
)

var configurationAPIKey = "0000"
var obsfucatorKey = []byte("0123456789")
var databaseName = "main.db"
var hostname = "http://127.0.0.1:8080"
var listenAddr = ":8080"

func shortenRoute(ctx iris.Context, session *durls.Session) {
	apiKey := ctx.URLParam("key")
	if apiKey != configurationAPIKey {
		ctx.StatusCode(401)
		return
	}
	url := ctx.URLParam("url")
	responseType := ctx.URLParamDefault("response_type", "plain_text")
	short := hostname + "/" + session.Shorten(url)
	if responseType == "json" {
		ctx.JSON(iris.Map{
			"action": "shorten",
			"result": short})
	} else if responseType == "plain_text" {
		ctx.Text(short)
	} else {
		ctx.StatusCode(400)
	}
}

func lookupRoute(ctx iris.Context, session *durls.Session) {
	apiKey := ctx.URLParam("key")
	if apiKey != configurationAPIKey {
		ctx.StatusCode(401)
		return
	}
	urlEnding := ctx.URLParam("url_ending")
	responseType := ctx.URLParamDefault("response_type", "plain_text")
	full, err := session.Lookup(urlEnding)
	if err != nil {
		ctx.StatusCode(404)
		return
	}
	if responseType == "json" {
		ctx.JSON(iris.Map{
			"action": "lookup",
			"result": full})
	} else if responseType == "plain_text" {
		ctx.Text(full)
	} else {
		ctx.StatusCode(400)
	}
}

func redirectRoute(ctx iris.Context, session *durls.Session) {
	url := ctx.Params().Get("url")
	full, err := session.Lookup(url)
	if err != nil {
		ctx.StatusCode(404)
		return
	}
	ctx.Redirect(full)
}

func main() {
	log.Println("Starting program...")
	session := durls.OpenSession(databaseName, obsfucatorKey)
	log.Println(fmt.Sprintf("Running on %v", hostname))
	app := iris.Default()
	app.Get("/api/v2/action/shorten", func(ctx iris.Context) {
		shortenRoute(ctx, session)
	})
	app.Get("/api/v2/action/lookup", func(ctx iris.Context) {
		lookupRoute(ctx, session)
	})
	app.Get("/{url: string regexp([a-z0-9]{6,7})}", func(ctx iris.Context) {
		redirectRoute(ctx, session)
	})
	app.Run(iris.Addr(listenAddr))
}
