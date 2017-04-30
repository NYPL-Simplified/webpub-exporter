package main

import (
	"encoding/json"
        "io/ioutil"
	"os"
        "strings"
        "time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/readium/r2-streamer-go/parser"
        "github.com/readium/r2-streamer-go/fetcher"
)

var (
	filenameFlag  = kingpin.Flag("file", "file to parse").Required().Short('f').String()
        outputDirFlag = kingpin.Flag("output", "directory to write output files to").Required().Short('o').String()
	urlFlag       = kingpin.Flag("url", "URL for the manifest").Short('u').String()
)

// Icon icon struct for AppInstall
type Icon struct {
	Src       string `json:"src"`
	Size      string `json:"size"`
	MediaType string `json:"type"`
}

// AppInstall struct for app install banner
type AppInstall struct {
	ShortName string `json:"short_name"`
	Name      string `json:"name"`
	StartURL  string `json:"start_url"`
	Display   string `json:"display"`
	Icons     []Icon   `json:"icons"`
}

func main() {

	kingpin.Version("0.0.1")
	kingpin.Parse()

        var outputDir = *outputDirFlag
        if !strings.HasSuffix(outputDir, "/") {
                outputDir += "/"
        }

        _ = os.Mkdir(outputDir, os.ModePerm)

	publication, _ := parser.Parse(*filenameFlag)

        _, err := publication.GetNavDoc()
        // Temporarily fill in HTML TOC the parser doesn't handle, until
        // the viewer can do something with publication.TOC.
        if err != nil {
                for i, link := range publication.Spine {
                        if link.Href == "toc-display.xhtml" {
                                publication.Spine[i].Rel = append(link.Rel, "contents")
                        }
                }
        }


        manifestJSON, _ := json.Marshal(publication)
        ioutil.WriteFile(outputDir + "/manifest.json", manifestJSON, 0644)

        cacheManifestString := "CACHE MANIFEST\n# timestamp "
        cacheManifestString += time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006")
        cacheManifestString += "\n\n"
        cacheManifestString += "index.html\n"
        cacheManifestString += "manifest.json\n"

        for _, link := range append(publication.Spine, publication.Resources...) {
                cacheManifestString += link.Href + "\n"
                var outputFile = outputDir + link.Href
                var parts = strings.Split(outputFile, "/")
                var outputFileDir = strings.Join(parts[:len(parts)-1], "/")
                _ = os.Mkdir(outputFileDir, os.ModePerm)
                fileReader, _, _ := fetcher.Fetch(&publication, link.Href)
                buff, _ := ioutil.ReadAll(fileReader)
                ioutil.WriteFile(outputFile, buff, 0644)
        }

        cacheManifestString += "\nNETWORK:\n*\n"
        ioutil.WriteFile(outputDir + "/manifest.appcache", []byte(cacheManifestString), 0644)

        var webManifest AppInstall
        webManifest.Display = "standalone"
        webManifest.StartURL = "index.html"
        webManifest.Name = publication.Metadata.Title.String()
        webManifest.ShortName = publication.Metadata.Title.String()

	cover, _ := publication.GetCover()
        var icon = Icon{
                Size: "144x144",
                Src: cover.Href,
                MediaType: cover.TypeLink,
        }
        webManifest.Icons = append(webManifest.Icons, icon)
        webManifestJSON, _ := json.Marshal(webManifest)
        ioutil.WriteFile(outputDir + "webapp.webmanifest", webManifestJSON, 0644)
}
