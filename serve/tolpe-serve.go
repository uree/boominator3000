package main

import (
    "net/http"
    "html/template"
    "fmt"
)


type ResponseData struct {
    Start string
    End   string
    Tolpe     []string
}

func main() {
    //http.Handle("/", http.FileServer(http.Dir("./templates")))
    tmpl := template.Must(template.ParseFiles("./templates/index.html"))
    http.HandleFunc("/tolpe", func(w http.ResponseWriter, r *http.Request) {

        query := r.URL.Query()
        start := query.Get("start")
        end := query.Get("end")

        fmt.Printf("params: ", start, end)

        response := ResponseData{
            Start: start,
            End: end,
            Tolpe: []string {
                "https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
                "https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
                "https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
            },
        }
        tmpl.Execute(w, response)
    })
    http.ListenAndServe(":8090", nil)
}
