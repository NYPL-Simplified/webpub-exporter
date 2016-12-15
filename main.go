package main

import (
	"encoding/json"
        "io/ioutil"
	"os"
        "strings"
        "time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/feedbooks/webpub-streamer/parser"
        "github.com/feedbooks/webpub-streamer/fetcher"
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
	Icons     Icon   `json:"icons"`
}

func main() {

	kingpin.Version("0.0.1")
	kingpin.Parse()

        var outputDir = *outputDirFlag
        if !strings.HasSuffix(outputDir, "/") {
                outputDir += "/"
        }

        _ = os.Mkdir(outputDir, os.ModePerm)

	publication := parser.Parse(*filenameFlag, *urlFlag)
        manifestJSON, _ := json.Marshal(publication)
        ioutil.WriteFile(outputDir + "/manifest.json", manifestJSON, 0644)

        cacheManifestString := "CACHE MANIFEST\n# timestamp "
        cacheManifestString += time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006")
        cacheManifestString += "\n\n"
        cacheManifestString += "index.html\n"

        for _, link := range append(publication.Spine, publication.Resources...) {
                cacheManifestString += link.Href + "\n"
                var outputFile = outputDir + link.Href
                var parts = strings.Split(outputFile, "/")
                var outputFileDir = strings.Join(parts[:len(parts)-1], "/")
                _ = os.Mkdir(outputFileDir, os.ModePerm)
                fileReader, _ := fetcher.Fetch(publication, link.Href)
                var fileBytes []byte
                fileReader.Read(fileBytes)
                ioutil.WriteFile(outputFile, fileBytes, 0644)
        }

        cacheManifestString += "\nNETWORK:\n*\n"
        ioutil.WriteFile(outputDir + "/manifest.appcache", []byte(cacheManifestString), 0644)

        var webManifest AppInstall
        webManifest.Display = "standalone"
        webManifest.StartURL = "index.html"
        webManifest.Name = publication.Metadata.Title
        webManifest.ShortName = publication.Metadata.Title

	cover := publication.GetCover()
        webManifest.Icons = Icon{
                Size: "144x144",
                Src: cover.Href,
                MediaType: cover.TypeLink,
        }
        webManifestJSON, _ := json.Marshal(webManifest)
        ioutil.WriteFile(outputDir + "webapp.webmanifest", webManifestJSON, 0644)
}
