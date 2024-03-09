package main

import (
	"net/http"

	"github.com/gladuz/ollama-chat-golang/handlers"
)

func main() {

	mux := http.NewServeMux()


	mux.HandleFunc("GET /chat", handlers.HandleChatIndex)
	mux.HandleFunc("GET /wschat", handlers.OpenSocketConn)
	mux.HandleFunc("GET /podsearch", handlers.HandlePodcastSearchIndex)
	mux.HandleFunc("POST /podsearch", handlers.HandlePodcastSearch)
	mux.HandleFunc("GET /podcast/{id}", handlers.HandlePodcastEpisodesShow)
	mux.HandleFunc("GET /episode/{id}", handlers.HandleEpisodeIndex)
	mux.HandleFunc("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))).ServeHTTP)
	mux.HandleFunc("GET /", handlers.HandleIndex)
	err := http.ListenAndServe(":4269", mux)
	if err != nil {
		panic(err)
	}
}	

// func prettyPrintResponse(messages []ModelResponse){
// 	for _, mes := range messages{
// 		if mes.Done{
// 			fmt.Printf("---- took %d seconds\n", mes.TotalDuration / 10^6)
// 		}
// 		fmt.Print(mes.Message.Content)
// 	}
// }

