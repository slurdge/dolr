package main

import (
	durls "durls/internal/durls"
	"github.com/kataras/iris"
	_ "github.com/kataras/iris/context"
	"log"
)

var configurationAPIKey = "0000"
var hostname = "http://127.0.0.1:8080"

func shortenRoute(ctx iris.Context) {
	apiKey := ctx.URLParam("key")
	if apiKey != configurationAPIKey {
		ctx.StatusCode(401)
		return
	}
	url := ctx.URLParam("url")
	responseType := ctx.URLParamDefault("response_type", "plain_text")
	log.Println(apiKey, url, responseType)
	short := hostname + "/" + durls.Shorten(url)
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

func lookupRoute(ctx iris.Context) {
	apiKey := ctx.URLParam("key")
	if apiKey != configurationAPIKey {
		ctx.StatusCode(401)
		return
	}
	urlEnding := ctx.URLParam("url_ending")
	responseType := ctx.URLParamDefault("response_type", "plain_text")
	full, err := durls.Lookup(urlEnding)
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

func redirectRoute(ctx iris.Context) {
	url := ctx.Params().Get("url")
	full, err := durls.Lookup(url)
	if err != nil {
		ctx.StatusCode(404)
		return
	}
	ctx.Redirect(full)
}

func main() {
	log.Println("Starting program...")
	durls.OpenDB("test.db")
	log.Println("Running on http://127.0.0.1:8080")
	app := iris.Default()
	app.Get("/api/v2/action/shorten", shortenRoute)
	app.Get("/api/v2/action/lookup", lookupRoute)
	app.Get("/{url: string regexp([a-z0-9]{6,7})}", redirectRoute)
	app.Run(iris.Addr("127.0.0.1:8080"))
}
