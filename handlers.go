package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

func execTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	files := []string{
		"./templs/base.tmpl",
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
	execTemplate(w, "./templs/index.tmpl", nil)
}

func HandlePodcastSearchIndex(w http.ResponseWriter, r *http.Request){
	execTemplate(w, "./templs/podsearch.tmpl", nil)
}

func HandlePodcastSearch(w http.ResponseWriter, r *http.Request){
	query := r.FormValue("query")
	if query == ""{
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}
	podSearchResult := parsePodcastSearchResult(query)
	execTemplate(w, "./templs/podresult.tmpl", podSearchResult)
}

func HandlePodcastEpisodesShow(w http.ResponseWriter, r *http.Request){
	id := r.PathValue("id")
	if id == ""{
		http.Error(w, "no id", http.StatusBadRequest)
		return
	}
	podcast := parsePodcastEpisodesByShow(id)
	execTemplate(w, "./templs/podcast.tmpl", podcast)
}



func HandleChat(w http.ResponseWriter, r *http.Request){
	messageForm := r.FormValue("message")

	if messageForm == ""{
		fmt.Println("no message")
		http.Error(w, "no message", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil{
		log.Fatal("conntection error", err)
	}
	ollamaRequestChan := make(chan ModelResponse)
	OllamaRequest(ollamaRequestChan, messageForm)
	var chatResponse ModelResponse
	for chatResponse = range ollamaRequestChan{
		if err != nil {
			log.Fatal("json encoding error on message", err)
		}
		fmt.Println(chatResponse.Message.Content)
		err = conn.WriteJSON(chatResponse)
		if err != nil {
			log.Println("write:", err)
		}
	}

}


func OpenSocketConn(w http.ResponseWriter, r *http.Request){
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
			if mt == CHAT_RESULT{
				ollamaRequestChan := make(chan ModelResponse)
				OllamaRequest(ollamaRequestChan, string(message))
				var chatResponse ModelResponse
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