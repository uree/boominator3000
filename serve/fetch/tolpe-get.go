package main

import (
	"strings"
    "fmt"
    "net/http"
	"html/template"
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

type ResponseData struct {
    Start string
    End   string
    Tolpe     []string
}

func between(start, end, check time.Time) bool {
    return check.After(start) && check.Before(end)
}

const
(
    RFC3339     = "2006-01-02T15:04Z"
	BASICDATE	= "2006-01-02"
)


func main() {
	//http.Handle("/", http.FileServer(http.Dir("./templates")))
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	http.HandleFunc("/tolpe", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		from := query.Get("from")
		to := query.Get("to")

		// testdata := []string {
		// 	"https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
		// 	"https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
		// 	"https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
		// }

		// an option in the future
		//https://stackoverflow.com/questions/37118281/dynamically-refresh-a-part-of-the-template-when-a-variable-is-updated-golang
		from_format := from + "T00:00Z"
		fmt.Printf("params: ", from_format, to)

		bclinks := fetchXML(from_format)
		fmt.Printf("bclinks: ", bclinks)



		response := ResponseData{
			Start: from,
			End: to,
			Tolpe: bclinks,
		}
		tmpl.Execute(w, response)
	})
	http.ListenAndServe(":8090", nil)

}

func fetchXML(from string)([]string) {
	start := time.Now()

	var tolpe[] string

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

	parsed := time.Since(start)
	fmt.Printf("Parsed [%s]", parsed)
	fmt.Printf("\n")

	c := make(chan string)

	timeFilter, _ := time.Parse(RFC3339, from)
	fmt.Println("timeFilter: ", timeFilter)

	go fetchBC(urlset, timeFilter, c)

	for bc := range c {
		fmt.Println("BANDCAMP: ", bc)
		tolpe = append(tolpe, bc)
		// fmt.Println("\n")
		elapsed := time.Since(start)
		fmt.Printf("Took [%s]", elapsed)
		fmt.Println("[ALL INNER]", tolpe)
	}
	//9 urls take 4s
	fmt.Println("[ALL]", tolpe)

	return tolpe

}

func fetchBC(urlset URLSet, timeFilter time.Time, c chan string) {

	for i := 0; i < len(urlset.URLSet); i++ {
		loc := strings.ToLower(urlset.URLSet[i].Location)
		mod, _ := time.Parse(RFC3339, urlset.URLSet[i].LastModifiedDate)

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
					fmt.Printf("iframe:\n %v", sauce)
					if strings.Contains(sauce, "//bandcamp") {
						c <- sauce
					}
				})
			}
		}
	}
	close(c)
}
