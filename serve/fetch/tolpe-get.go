package main

import (
	"strings"
    "fmt"
    "net/http"
    "encoding/xml"
	"io/ioutil"
	"time"
	//"sync"

	"github.com/PuerkitoBio/goquery"
)


type URLSet struct {
	XMLName xml.Name `xml:urlset`
    URLSet	[]SitemapURL `xml:"url"`
}

type SitemapURL struct {
	XMLName	         xml.Name `xml:"url"`
    Location         string `xml:"loc"`
    LastModifiedDate string `xml:"lastmod"`
    ChangeFrequency  string `xml:"changefreq"`
}


func inTimeSpan(start, end, check time.Time) bool {
    return check.After(start) && check.Before(end)
}

const
(
    RFC3339     = "2006-01-02T15:04Z"
)


func main() {
	start := time.Now()

    resp, err := http.Get("http://radiostudent.si/sitemap.xml?page=1")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)

	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	var urlset URLSet
    xml.Unmarshal(byteValue, &urlset)

	c := make(chan string)
	//cFin := make(chan bool)

	timeFilter, _ := time.Parse(RFC3339, "2021-10-01T15:00Z")
	fmt.Println("timeFilter: ", timeFilter)

	// var wg sync.WaitGroup
	// wg.Add(1)
	//
	// go func() {
	// 	fetchBC(urlset, timeFilter, c)
	// 	wg.Done()
	//
	// }()
	//
	// //blocks until counter is 0
	// wh.Wait()

	// for elem := range c {
    //     fmt.Println("BANDCAMP: ", elem, '\n')
    // }

	// bandcamp := <- c
	// fmt.Println("\nbandcamp: ", bandcamp)

	//
	// bandcamps[bandcamp] = true
	//
	// // bandcamps = append(bandcamps, bandcamp)
	//fmt.Println("bandcamps: ", bandcamps)

	for i := 0; i < len(urlset.URLSet); i++ {
		loc := strings.ToLower(urlset.URLSet[i].Location)
		mod, _ := time.Parse(RFC3339, urlset.URLSet[i].LastModifiedDate)
		go fetchBC(i, loc, mod, timeFilter, c)

	}

	for bc := range c {
		fmt.Println("BANDCAMP: ", bc)
		fmt.Println("\n")
		elapsed := time.Since(start)
		fmt.Printf("Took [%s]", elapsed)
	}
	//9 urls take 4s





}

func fetchBC(i int, loc string, mod time.Time, timeFilter time.Time, c chan string) {

	if strings.Contains(loc, "/glasba/tolpa-bumov/"){
		if mod.After(timeFilter) {
			fmt.Println("Location: " + loc)
			//fmt.Println("Modified: " + mod)
			resp, err := http.Get(loc)

			if err != nil {
			  c <- fmt.Sprint(err) // send to channel ch
			  return
			}

			defer resp.Body.Close()

			fmt.Println("Response status:", resp.Status)

			doc, err := goquery.NewDocumentFromReader(resp.Body)

			if err != nil {
				fmt.Printf("error: %v", err)
			}

			//ignore yt iframes yo!

			doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
				sauce, _ := s.Attr("src")
				//fmt.Printf("iframe:\n %v", sauce)
				c <- sauce
			})
		}
	}
}
