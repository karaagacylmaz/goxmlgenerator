package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"encoding/xml"
	"io"

	"github.com/go-pg/pg/v9"
	_ "github.com/joho/godotenv/autoload"
)

var (
	port     string
	host     string
	user     string
	database string
	password string
	storage  string
)

type (
	Link struct {
		tableName  struct{} `pg:"sitemap_links"`
		Slug       string
		ImageUrls  []string `pg:",array"`
		ChangeType string
	}

	Url struct {
		XMLName    xml.Name  `xml:"url"`
		Loc        string    `xml:"loc"`
		Lastmod    time.Time `xml:"lastmod"`
		Changefreq string    `xml:"changefreq"`
		MyImages   *Image
	}

	Image struct {
		XMLName xml.Name `xml:"image:image,omitempty"`
		Loc     []string `xml:"image:loc,omitempty"`
	}

	Urlset struct {
		XMLName    xml.Name `xml:"urlset"`
		Urls       []Url
		Xmlns      string `xml:"xmlns,attr"`
		XmlnsImage string `xml:"xmlns:image,attr"`
	}
)

func init() {
	host = os.Getenv("DB_HOST")
	port = os.Getenv("DB_PORT")
	user = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	database = os.Getenv("DB_NAME")
	if s := os.Getenv("DATA_PATH"); s != "" {
		storage = s
	} else {
		storage = "./"
	}
}

func main() {
	start := time.Now()
	db := pg.Connect(&pg.Options{
		User:     user,
		Password: password,
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Database: database,
	})
	defer db.Close()

	var ls []Link
	err := db.Model(&ls).Select()
	fmt.Println(ls[0])
	if err != nil {
		panic(err)
	}

	v := &Urlset{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", XmlnsImage: "http://www.google.com/schemas/sitemap-image/1.1"}

	for index, _ := range ls {
		slug, _ := url.Parse(ls[index].Slug)
		images := ls[index].ImageUrls
		changeFreq := ls[index].ChangeType
		u := Url{
			Loc: slug.String(), Lastmod: time.Now().UTC(), Changefreq: changeFreq,
		}
		if len(images) != 0 {
			u.MyImages = &Image{Loc: images}
		}
		v.Urls = append(v.Urls, u)
	}

	file, _ := os.Create(filepath.FromSlash(storage) + "/sitemap.xml")
	defer file.Close()

	xmlWriter := io.Writer(file)
	enc := xml.NewEncoder(xmlWriter)
	enc.Indent(" ", "  ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}

	defer fmt.Printf("%s ready!\nElapsed %s\n", "Sitemap", time.Since(start))
}
