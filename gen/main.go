package gen

import (
	"net/http"
	"strings"

	"github.com/fossunited/ograph-gen/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-rat/chix"
	"github.com/mskrha/svg2png"
)

type GenResource struct {
	RoutesConfig *[]utils.RouteConfig
	SVGConverter *svg2png.Converter
}

func (rs GenResource) Routes() chi.Router {
	r := chi.NewRouter()
	for _, route := range *rs.RoutesConfig {
		r.Get("/"+route.Name, func(w http.ResponseWriter, r *http.Request) {
			params := make(map[string]string)
			bind := chix.NewBind(r)
			defer bind.Release()
			if err := bind.Query(params); err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Could not parse query strings: " + err.Error()))
				return
			}

			svgStr := string(route.SVG)
			for _, val := range route.Replacements {
				svgStr = strings.ReplaceAll(svgStr, val.ReplacementValue, params[val.ReplacementName])
			}

			png, err := rs.SVGConverter.Convert([]byte(svgStr))
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
