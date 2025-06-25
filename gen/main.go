package gen

import (
	"net/http"
	"bytes"

	"github.com/fossunited/ograph-gen/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-rat/chix"
	"github.com/mskrha/svg2png"
	"text/template"
)

type GenResource struct {
	RoutesConfig *[]utils.RouteConfig
	SVGConverter *svg2png.Converter
}

func (rs GenResource) Routes() chi.Router {
	r := chi.NewRouter()
	for i, _ := range *rs.RoutesConfig {
		conf := (*rs.RoutesConfig)[i]
		r.Get("/"+conf.Name, func(w http.ResponseWriter, r *http.Request) {
			params := make(map[string]string)
			bind := chix.NewBind(r)
			defer bind.Release()
			if err := bind.Query(params); err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Could not parse query strings: " + err.Error()))
				return
			}

			svgStr := string(conf.SVG)
			tmpl, err := template.New("svg").Parse(svgStr)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Unable to parse SVG: " + err.Error()))
				return
			}
			
			svgBuf := bytes.NewBuffer([]byte(""))
			err = tmpl.Execute(svgBuf, params)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Unable to parse SVG: " + err.Error()))
				return
			}

			png, err := rs.SVGConverter.Convert(svgBuf.Bytes())
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Could not convert svg to png: " + err.Error()))
				return
			}

			w.Header().Set("Content-Type", "image/png")
			w.Write(png)
		})
	}
	return r
}
