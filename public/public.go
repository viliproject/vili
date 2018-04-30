package public

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
)

// Assets is the variable that stores assets
var Assets AssetsWrapper

// AssetsWrapper is a wrapper for static assets
type AssetsWrapper struct {
	Hash string
}

// Stats is the json object for webpack stats
type Stats struct {
	BuildDir string
	Hash     string `json:"hash"`
}

var stats Stats
var buildDir string

// LoadStats loads the Assets variable
func LoadStats(d string) error {
	buildDir = d
	statsBytes, err := ioutil.ReadFile(path.Join(d, "stats.json"))
	if err != nil {
		return err
	}
	json.Unmarshal(statsBytes, &stats)
	stats.BuildDir = buildDir
	return nil
}

// GetHash returns the hash of the built assets
func GetHash(reload bool) string {
	if reload {
		LoadStats(buildDir)
	}
	return stats.Hash
}

// StaticHandler serves static files
func StaticHandler(c echo.Context) error {
	name := c.Param("name")
	if strings.HasPrefix(name, ".") {
		return c.String(http.StatusNotFound, "Not Found")
	}
	// var contentType string
	// if strings.HasSuffix(name, ".js") {
	// 	contentType = "application/javascript"
	// } else if strings.HasSuffix(name, ".js.map") {
	// 	contentType = "application/json"
	// } else {
	// 	return c.String(http.StatusNotFound, "Not Found")
	// } // TODO
	mimetype := mime.TypeByExtension(filepath.Ext(name))
	c.Response().Header().Set("Content-Type", mimetype)
	http.ServeFile(c.Response(), c.Request(), path.Join(stats.BuildDir, name))
	return nil
}
