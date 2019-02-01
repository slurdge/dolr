package main

import (
	dolr "dolr/internal/dolr"
	"fmt"
	"log"

	"github.com/kataras/iris"
	_ "github.com/kataras/iris/context"
	"github.com/tkanos/gonfig"
)

type configuration struct {
	APIKey        string `env:"API_KEY"`
	ObsfucatorKey string `env:"OBS_KEY"`
	Database      string `env:"DATABASE"`
	Hostname      string `env:"HOSTNAME"`
	ListenAddr    string `env:"LISTEN_ADDR"`
	UseTLS        bool
	SslKeyFile    string
	SslCertFile   string
}

func shortenRoute(ctx iris.Context, session *dolr.Session, apiKey string, hostname string) {
	apiKeyParam := ctx.URLParam("key")
	if apiKey != apiKeyParam {
		ctx.StatusCode(401)
		return
	}
	url := ctx.URLParam("url")
	responseType := ctx.URLParamDefault("response_type", "plain_text")
	short, err := session.Shorten(url)
	if err != nil {
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
	ctx.StatusCode(400)
}

func lookupRoute(ctx iris.Context, session *dolr.Session, apiKey string) {
	apiKeyParam := ctx.URLParam("key")
	if apiKey != apiKeyParam {
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

func redirectRoute(ctx iris.Context, session *dolr.Session) {
	url := ctx.Params().Get("url")
	full, err := session.Lookup(url)
	if err != nil {
		ctx.StatusCode(404)
		return
	}
	ctx.Redirect(full)
}

func postRoute(ctx iris.Context, session *dolr.Session, hostname string) {
	url := ctx.PostValue("url")
	short, err := session.Shorten(url)
	if err == nil {
		ctx.ViewData("shortened", true)
		ctx.ViewData("full", url)
		ctx.ViewData("short", hostname+"/"+short)
	} else {
		ctx.ViewData("shortened", false)
	}
	ctx.View("index.html")
}

func main() {
	log.Println("Starting program...")
	configuration := configuration{
		APIKey:        "0123456789",
		Database:      "main.db",
		ObsfucatorKey: "0123456789",
		ListenAddr:    "127.0.0.1:8080",
		Hostname:      "http://localhost:8080",
		UseTLS:        false,
		SslCertFile:   "",
		SslKeyFile:    "",
	}
	err := gonfig.GetConf("dolr.json", &configuration)
	if err != nil {
		panic(err)
	}

	session := dolr.OpenSession(configuration.Database, []byte(configuration.ObsfucatorKey))
	log.Println(fmt.Sprintf("Running on %v", configuration.Hostname))
	app := iris.Default()
	app.Get("/api/v2/action/shorten", func(ctx iris.Context) {
		shortenRoute(ctx, session, configuration.APIKey, configuration.Hostname)
	})
	app.Get("/api/v2/action/lookup", func(ctx iris.Context) {
		lookupRoute(ctx, session, configuration.APIKey)
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
	app.Post("/", func(ctx iris.Context) {
		postRoute(ctx, session, configuration.Hostname)
	})
	if !configuration.UseTLS {
		app.Run(iris.Addr(configuration.ListenAddr))
	} else {
		app.Run(iris.TLS(configuration.ListenAddr, configuration.SslCertFile, configuration.SslKeyFile))
	}

}
