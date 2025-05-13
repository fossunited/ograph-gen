package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/docgen"
	"github.com/mskrha/svg2png"

	"github.com/fossunited/ograph-gen/gen"
	"github.com/fossunited/ograph-gen/utils"
)

func main() {
	conf, err := utils.ConfigDecode()
	if err != nil {
		log.Fatal("unable to parse config.json: " + err.Error())
	}

	routes, err := utils.LoadRoutes(conf)
	if err != nil {
		log.Fatal("unable to load routes: " + err.Error())
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/ping"))

	// Every sub-route has access to so it can do whatever the fuck it needs to
	r.Mount("/gen", gen.GenResource{RoutesConfig: &routes, SVGConverter: svg2png.New()}.Routes())

	// Generate docs using docgen
	// TODO: MAKE THIS INTO A SEPARATE FILE OR SERVE IT
	fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
		ProjectPath: "github.com/go-chi/chi/v5",
		Intro:       "Welcome to the chi/_examples/rest generated docs.",
	}))

	http.ListenAndServe(conf.Host, r)
}
