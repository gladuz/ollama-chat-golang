package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
const(
	CHAT_RESULT = 1
)

const (
	CHAT_URL       = "http://localhost:11434/api/chat"
	ROLE_ASSISTANT = "assistant"
	ROLE_USER      = "user"
	maxBufferSize = 512 * 1024
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Images string `json:"images,omitempty"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ModelResponse struct {
	Model              string      `json:"model"`
	CreatedAt          time.Time   `json:"created_at"`
	Message            ChatMessage `json:"message"`
	Done               bool        `json:"done"`
	TotalDuration      int64       `json:"total_duration"`
	LoadDuration       int64       `json:"load_duration"`
	PromptEvalCount    int         `json:"prompt_eval_count"`
	PromptEvalDuration int64       `json:"prompt_eval_duration"`
	EvalCount          int         `json:"eval_count"`
	EvalDuration       int64       `json:"eval_duration"`
}

func OllamaRequest(c chan ModelResponse, message string){
	newChat := ChatRequest{
		Model: "mistral",
		Messages: []ChatMessage{
			{Role: ROLE_USER, Content: message},
		},
	}
	go PostChatRequest(newChat, c)
}

func PostChatRequest(newChat ChatRequest, c chan ModelResponse) {
	requestContent, err := json.Marshal(&newChat)
	if err != nil {
		log.Fatal("json encoding error on message")
	}
	postResponse, err := http.Post(CHAT_URL, "application/json", bytes.NewBuffer(requestContent))
	if err != nil {
		log.Fatal("Request error", err)
	}

	defer postResponse.Body.Close()

	scanner := bufio.NewScanner(postResponse.Body)
	// increase the buffer size to avoid running out of space
	scanBuf := make([]byte, 0, maxBufferSize)
	scanner.Buffer(scanBuf, maxBufferSize)
	for scanner.Scan() {
		var errorResponse struct {
			Error string `json:"error,omitempty"`
		}

		bts := scanner.Bytes()
		if err := json.Unmarshal(bts, &errorResponse); err != nil {
			log.Fatal("unmarshal error:", err)
		}
		var chatResponse ModelResponse
		if err := json.Unmarshal(bts, &chatResponse); err != nil{
			log.Fatal("error unmarshalling:", err)
		}
		c <- chatResponse
	}
	close(c)
}

