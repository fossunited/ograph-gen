package utils

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type RouteConfig struct {
	Name         string        `json:"name"`
	SVG          []byte        `json:"svg"`
}

type Config struct {
	Host    string   `json:"host"`
	DataDir string   `json:"datadir"`
	Routes  []string `json:"routes"`
}

func ConfigDecode() (Config, error) {
	confByte, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("could not open config.json: " + err.Error())
	}

	var conf Config
	json.Unmarshal(confByte, &conf)

	if conf.Host == "" {
		conf.Host = ":3333"
	}
	if conf.DataDir == "" {
		return Config{}, errors.New("data dir not declared in config")
	}
	if len(conf.Routes) == 0 {
		return Config{}, errors.New("no routes declared in config")
	}
	return conf, nil
}

func LoadRoutes(conf Config) ([]RouteConfig, error) {
	var routes []RouteConfig
	for _, route := range conf.Routes {
		var routeConf RouteConfig
		var err error

		routeConf.Name = route
		routeConf.SVG, err = os.ReadFile(conf.DataDir + "/" + route + ".svg") // Read the route's svg
		if err != nil {
			return []RouteConfig{}, err
		}
		routes = append(routes, routeConf)
	}
	return routes, nil
}
