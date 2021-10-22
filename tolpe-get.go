package main

import (
	"strings"
    "fmt"
    "net/http"
	"html/template"
    "encoding/xml"
	"io/ioutil"
	"time"
	"os"
	"strconv"
	"io"
	//"sync"

	"github.com/PuerkitoBio/goquery"
)

// NOTES

// perhaps useful for catching the event which prompts bandcamp to sent analytics data phase = complete
// https://bandcamp.com/stat_record?kind=track+play&track_id=2154565882&track_license_id=&from=embedded+album+player&from_url=http%3A%2F%2Flocalhost%3A8090%2F&stream_duration=214.748962&phase=complete&reference_num=743361424&rand=36982917390818215
// https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/Intercept_HTTP_requests

//events
//https://bluerivermountains.com/en/log-all-javascript-events
//monitorEvents(window,event); (in chrome .. no events though)
// no related events
// but the answer must be in the embedded-player.js file
// this could be useful https://github.com/lovethebomb/bandcamp-feed-playlist
//This is a lazy Chrome Extension that adds a mini player on the Bandcamp feed page. It allows you to quickly Play/Pause, go to next and previous track, and support autoplay of the tracks shared on your feed.

// embedded_player ln 1563 Stats && Stats.PhasedStat
// window[0].Player.PlayStat

// supposedly unofficial apis exist
// https://bandcamp.com/api/discover/3/get_web?g=all&gn=0&p=0&s=top&f=all&w=0&callback=jQuery35107982526342886465_1634893255270&_=1634893255271
// how to get data (like file for streaming)
// https://nevolin.be/bandcamp/main.js?_=1634893255269
// https://bandcamp.com/api/hub/2/dig_deeper
//https://github.com/michaelherger/Bandcamp-API not very useful



// DATA STRUCTURES

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

type SitemapSet struct {
	XMLName xml.Name `xml:sitemapindex`
    SSet	[]SitemapSitemap `xml:"sitemap"`
}

type SitemapSitemap struct {
	XMLName	         xml.Name `xml:"sitemap"`
    Location         string `xml:"loc"`
    LastModifiedDate string `xml:"lastmod"`
    ChangeFrequency  string `xml:"changefreq"`
}

type ResponseData struct {
    Start string
    End   string
    Tolpe     []string
}

// HELPER FUNCTIONS

func between(start, end, check time.Time) bool {
    return check.After(start) && check.Before(end)
}

func openFile(filepath string)(URLSet) {
	xmlFile, err := os.Open(filepath)

	if err != nil {
		fmt.Println(err)
	}

	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var urlset URLSet
	xml.Unmarshal(byteValue, &urlset)
	return urlset
}

func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}


// head does not contain this
func getSize(url string) {
	res, err := http.Head(url)
	if err != nil {
		panic(err)
	}
	contentlength:=res.ContentLength
	fmt.Printf("ContentLength:%v",contentlength)
	//return contentlength
}

const
(
    RFC3339     = "2006-01-02T15:04Z"
	BASICDATE	= "2006-01-02"
)


// MAIN

func main() {
	//http.Handle("/", http.FileServer(http.Dir("./templates")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	http.HandleFunc("/tolpe", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		from := query.Get("from")
		to := query.Get("to")
		refresh := query.Get("update")

		update := false

		if to == "" {
			ct := time.Now()
			to = ct.Format("2006-01-02")
		}

		if refresh == "true" {
			update = true
		} else {
			fmt.Println("<UPDATE OFF>")
		}

		// testdata := []string {
		// 	"https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
		// 	"https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
		// 	"https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
		// }

		// an option in the future
		//https://stackoverflow.com/questions/37118281/dynamically-refresh-a-part-of-the-template-when-a-variable-is-updated-golang
		from_format := from + "T00:00Z"
		to_format := to + "T00:00Z"
		fmt.Printf("[PARAMS] ", from_format, to_format)

		// one month takes 13s
		bclinks := fetchXML(from_format, to_format, update)
		fmt.Printf("[COUNT] ", len(bclinks))

		response := ResponseData{
			Start: from,
			End: to,
			Tolpe: bclinks,
		}
		tmpl.Execute(w, response)
	})
	http.ListenAndServe(":8090", nil)

}

// CORE FUNCTIONS

func fetchXML(from string, to string, update bool)([]string) {
	start := time.Now()

	var tolpe[] string

	sitemaps := []string {"http://radiostudent.si/sitemap.xml?page=1","http://radiostudent.si/sitemap.xml?page=2"}

	// check if sitemap has changed since last call && download
	cfetch := make(chan string)

	if update {
		go fetchSitemaps(sitemaps, cfetch)
	}

	/// else just open the files directly (will this be slower? naah)
	sitemap1 := openFile("./temp/sitemap1.xml")
	sitemap2 := openFile("./temp/sitemap2.xml")


	parsed := time.Since(start)
	fmt.Printf("Parsed [%s]", parsed)
	fmt.Printf("\n")


	c := make(chan string)
	c2 := make(chan string)

	from_date, _ := time.Parse(RFC3339, from)
	to_date := start

	if to != "" {
		to_date, _ = time.Parse(RFC3339, to)
	}

	fmt.Println("timeFilter: ", from_date, to_date)

	fmt.Println("--- goroutine sitemap 1 start --- ")
	go fetchBC(sitemap1, from_date, to_date, c)
	fmt.Println("--- goroutine sitemap 2 start --- ")
	go fetchBC(sitemap2, from_date, to_date, c2)

	for bc := range c {
		//fmt.Println("BANDCAMP1: ", bc)
		tolpe = append(tolpe, bc)
	}

	for bc := range c2 {
		//fmt.Println("BANDCAMP2: ", bc)
		tolpe = append(tolpe, bc)
	}

	elapsed := time.Since(start)
	fmt.Println("Took [%s]", elapsed)

	fmt.Println("[ALL]", tolpe)

	return tolpe
}

func fetchSitemaps(sitemaps []string, c chan string) {
	for i := 0; i < len(sitemaps); i++ {
		filename := "./temp/sitemap"+strconv.Itoa(i)+".xml"
		//url := sitemaps[i]
		downloadFile(filename, sitemaps[i])
	}
	close(c)
}

func fetchBC(urlset URLSet, from_date time.Time, to_date time.Time, c chan string) {

	for i := 0; i < len(urlset.URLSet); i++ {
		loc := strings.ToLower(urlset.URLSet[i].Location)
		mod, _ := time.Parse(RFC3339, urlset.URLSet[i].LastModifiedDate)

		if strings.Contains(loc, "/glasba/tolpa-bumov/"){
			if between(from_date, to_date, mod) {
				//fmt.Println("Location: " + loc)
				//fmt.Println("Modified: " + mod)
				resp, err := http.Get(loc)

				if err != nil {
				  c <- fmt.Sprint(err) // send to channel ch
				  return
				}

				defer resp.Body.Close()

				//fmt.Println("Response status:", resp.Status)

				doc, err := goquery.NewDocumentFromReader(resp.Body)

				if err != nil {
					fmt.Printf("error: %v", err)
				}

				//ignore yt iframes yo!

				doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
					sauce, _ := s.Attr("src")
					//fmt.Printf("iframe:\n %v", sauce)
					if strings.Contains(sauce, "//bandcamp") {
						sauce = strings.Replace(sauce, "/tracklist=false", "tracklist=true", -1)
						c <- sauce
					}
				})
			}
		}
	}
	close(c)
}
