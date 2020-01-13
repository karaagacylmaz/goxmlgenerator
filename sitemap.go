package main

import (
	"fmt"
	"net/url"
	"os"
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
		MyImages   Image
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
}

func main() {
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
		//, MyImages: Image{Loc: images}
		changeFreq := ls[index].ChangeType
		v.Urls = append(v.Urls, Url{Loc: slug.String(), Lastmod: time.Now().UTC(), Changefreq: changeFreq, MyImages: Image{Loc: images}})
	}

	filename := "sitemap.xml"
	file, _ := os.Create(filename)

	xmlWriter := io.Writer(file)
	enc := xml.NewEncoder(xmlWriter)
	enc.Indent(" ", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
