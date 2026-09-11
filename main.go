package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/docgen"

	"github.com/fossunited/ograph-gen/convert"
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

	converter := convert.New()
	if path, err := exec.LookPath("rsvg-convert"); err == nil {
		converter.SetBinary(path)
	}

	r.Mount("/gen", gen.GenResource{RoutesConfig: &routes, SVGConverter: converter}.Routes())

	// Generate docs using docgen
	// TODO: MAKE THIS INTO A SEPARATE FILE OR SERVE IT
	fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
		ProjectPath: "github.com/go-chi/chi/v5",
		Intro:       "Welcome to the chi/_examples/rest generated docs.",
	}))

	http.ListenAndServe(conf.Host, r)
}
