package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/a-h/templ"
	"github.com/gladuz/ollama-chat-golang/db"
	"github.com/gladuz/ollama-chat-golang/llm"
	"github.com/gladuz/ollama-chat-golang/ui"
	"github.com/gorilla/websocket"
)

func execTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	files := []string{
		".ui/templs/base.tmpl",
		tmpl,
	}

	// Use the template.ParseFiles() function to read the files and store the
	// templates in a template set. Notice that we use ... to pass the contents
	// of the files slice as variadic arguments.
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleIndex(w http.ResponseWriter, r *http.Request){

	templ.Handler(ui.Index()).ServeHTTP(w, r)
	//execTemplate(w, ".ui/templs/index.tmpl", nil)
}

func HandlePodcastSearchIndex(w http.ResponseWriter, r *http.Request){
	templ.Handler(ui.PodcastSearch()).ServeHTTP(w, r)
}

func HandlePodcastSearch(w http.ResponseWriter, r *http.Request){
	query := r.FormValue("query")
	if query == ""{
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}
	podSearchResult := db.ParsePodcastSearchResult(query)
	templ.Handler(ui.PodcastSearchResult(podSearchResult)).ServeHTTP(w, r)
}

func HandlePodcastEpisodesShow(w http.ResponseWriter, r *http.Request){
	id := r.PathValue("id")
	if id == ""{
		http.Error(w, "no id", http.StatusBadRequest)
		return
	}
	podcast := db.ParsePodcastEpisodesByShow(id)
	templ.Handler(ui.PodcastEpisodesShow(podcast)).ServeHTTP(w, r)
}

func HandleEpisodeIndex(w http.ResponseWriter, r *http.Request){
	id := r.PathValue("id")
	if id == ""{
		http.Error(w, "no id", http.StatusBadRequest)
		return
	}
	episode := db.ParseEpisode(id)
	templ.Handler(ui.EpisodeIndex(episode)).ServeHTTP(w, r)
}

func HandleChatIndex(w http.ResponseWriter, r *http.Request){
	templ.Handler(ui.ChatIndex()).ServeHTTP(w, r)
}


func OpenSocketConn(w http.ResponseWriter, r *http.Request){
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil{
		log.Fatal("connection error", err)
	}

	defer conn.Close()
	go func() {
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Println(string(message))
			log.Printf("recv: %s, mt: %d", message, mt)
			if mt == llm.CHAT_RESULT{
				ollamaRequestChan := make(chan llm.ModelResponse)
				llm.OllamaRequest(ollamaRequestChan, string(message))
				var chatResponse llm.ModelResponse
				for chatResponse = range ollamaRequestChan{
					if err != nil {
						log.Fatal("json encoding error on message", err)
					}
					err = conn.WriteJSON(chatResponse)
					if err != nil {
						log.Println("write:", err)
					}
				}
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	//keep the connection alive
	//close the connection if the client disconnects
	//or if the server is interrupted
	for {
		select {
		case <-ticker.C:
			
		case <-interrupt:
			log.Println("interrupt")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			time.Sleep(time.Second)
			conn.Close()
			os.Exit(0)
		}

	}
}