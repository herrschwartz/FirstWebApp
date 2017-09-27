package static

import (
  "fmt"
  "html/template"
  "net/http"
  "flag"
  "log"
  "time"
  "appengine"
  "appengine/datastore"
	"appengine/user"

  "code.google.com/p/google-api-go-client/youtube/v3"
  "code.google.com/p/google-api-go-client/googleapi/transport"	 
)

type Post struct{
	Author string
	Content string
	Date time.Time
}

func init() {
	http.HandleFunc("/",static)
	http.HandleFunc("/checkform", root)
        http.HandleFunc("/formresult", upper)
	http.HandleFunc("/utube", utube)
	http.HandleFunc("/forum", forum)
	http.HandleFunc("/sign", sign)

}
func forumKey(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "Post", "default_post", 0, nil)
}

func forum(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Post").Ancestor(forumKey(c)).Order("-Date").Limit(10)
	posts := make([]Post, 0, 10)
	if _, err := q.GetAll(c, &posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := Template.Execute(w, posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
var Template = template.Must(template.New("book").Parse(`
<html>
  <head>
    <title>Forum Posts</title>
  </head>
  <body>
	<h1> Posts </h1>

    {{range .}}
    <div style = "padding: 8px; margin-bottom: 6px; margin-top 4px; background-color: #B6D4CD;border:1px">
      {{with .Author}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>An anonymous person wrote:</p>
      {{end}}
      <pre>{{.Content}}</pre>
    <a href = "#">Delete</a>
    <a href = "#">Edit</a>
    </div>
    {{end}}

    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Make Post"></div>
    </form>
  </body>
</html>
`))

func sign(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := Post{
		Content: r.FormValue("content"),
		Date:    time.Now(),
	}
	if u := user.Current(c); u != nil {
		g.Author = u.String()
	}
	key := datastore.NewIncompleteKey(c, "Post", forumKey(c))
	_, err := datastore.Put(c, key, &g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/forum", http.StatusFound)
}

func static(w http.ResponseWriter, r *http.Request){
	    c := appengine.NewContext(r)
    	    u := user.Current(c)
    if u == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
	http.ServeFile(w,r,"public/"+r.URL.Path)
	
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, rootForm)
}

var (
	query      = flag.String("query", "Google", "Search term")
	maxResults = flag.Int64("max-results", 25, "Max YouTube results")
)

const developerKey = "AIzaSyB8eGl8oN_ZZyGxTHM7QqJ6b_OhNkKar1Y"

func utube(w http.ResponseWriter, r *http.Request){
	flag.Parse()

	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(*query).
		MaxResults(*maxResults)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making search API call: %v", err)
	}

	// Group video, channel, and playlist results in separate lists.
	videos := make(map[string]string)
	channels := make(map[string]string)
	playlists := make(map[string]string)

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = item.Snippet.Title
		case "youtube#channel":
			channels[item.Id.ChannelId] = item.Snippet.Title
		case "youtube#playlist":
			playlists[item.Id.PlaylistId] = item.Snippet.Title
		}
	}

	printIDs("Videos", videos)
	printIDs("Channels", channels)
	printIDs("Playlists", playlists)
}

// Print the ID and title of each result in a list as well as a name that
// identifies the list. For example, print the word section name "Videos"
// above a list of video search results, followed by the video ID and title
// of each matching video.
func printIDs(sectionName string, matches map[string]string) {
	fmt.Printf("%v:\n", sectionName)
	for id, title := range matches {
		fmt.Printf("[%v] %v\n", id, title)
	}
	fmt.Printf("\n\n")
}

const rootForm = `
  <!DOCTYPE html>
    <html>
      <head>
        <meta charset="utf-8">
        <link rel="stylesheet" href="public/assets/css/main.css">
        <title>Name Checker</title>
      </head>
      <body>
        <h1>String Validator</h1>
        <p>Enter the correct name</p>
        <form action="/formresult" method="post" accept-charset="utf-8">
	  <input type="text" name="str" placeholder="Type a string..." id="str">
	  <input type="submit" value="Validate">
        </form>
      </body>
    </html>
`
var upperTemplate = template.Must(template.New("upper").Parse(upperTemplateHTML))

func upper(w http.ResponseWriter, r *http.Request) {
	strEntered := r.FormValue("str")
	if strEntered == "Tim"{
      err := upperTemplate.Execute(w, "You got my name right!")
      if err !=nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
      }
    } else {
      err := upperTemplate.Execute(w, "Thats not my name!")
      if err !=nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
      }
    }

}

const upperTemplateHTML = `
<!DOCTYPE html>
  <html>
    <head>
      <meta charset="utf-8">
      <link rel="stylesheet" href="public/assets/css/main.css">
      <title>String Upper Results</title>
    </head>
    <body>
      <pre>{{.}}</pre>
    </body>
  </html>
`
