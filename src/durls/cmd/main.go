package main

import (
	durls "durls/internal/durls"
	"fmt"
	"log"
	"strings"

	"github.com/kataras/iris"
	_ "github.com/kataras/iris/context"
)

var configurationAPIKey = "0000"
var obsfucatorKey = []byte("0123456789")
var databaseName = "main.db"
var hostname = "https://durls.test:8080"
var listenAddr = "localhost:8080"
var useTLS = true
var sslKeyFile = "./durls.test-key.pem"
var sslCertFile = "./durls.test.pem"

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
	tmpl := iris.HTML("./templates", ".html").Reload(true)
	app.RegisterView(tmpl)
	app.StaticWeb("/", "./static")
	indexHandler := func(ctx iris.Context) {
		ctx.ViewData("shortened", false)
		ctx.View("index.html")
	}
	app.Get("/", indexHandler)
	postHandler := func(ctx iris.Context) {
		url := ctx.PostValue("url")
		if !strings.HasPrefix(url, "http") || strings.TrimSpace(url) == "" {
			ctx.ViewData("shortened", false)
		} else {
			short := hostname + "/" + session.Shorten(url)
			ctx.ViewData("shortened", true)
			ctx.ViewData("full", url)
			ctx.ViewData("short", short)
		}
		ctx.View("index.html")
	}
	app.Post("/", postHandler)
	if !useTLS {
		app.Run(iris.Addr(listenAddr))
	} else {
		app.Run(iris.TLS(listenAddr, sslCertFile, sslKeyFile))
	}

}
