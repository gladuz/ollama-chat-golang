package main

import (
	"net/http"
)

func main() {

	mux := http.NewServeMux()


	mux.HandleFunc("GET /", HandleIndex)
	mux.HandleFunc("GET /chat", OpenSocketConn)
	mux.HandleFunc("POST /chat", HandleChat)
	mux.HandleFunc("GET /podsearch", HandlePodcastSearchIndex)
	mux.HandleFunc("POST /podsearch", HandlePodcastSearch)
	mux.HandleFunc("GET /podcast/{id}", HandlePodcastEpisodesShow)
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

