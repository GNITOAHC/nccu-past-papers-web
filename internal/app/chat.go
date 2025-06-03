package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"past-papers-web/templates"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type ChatHis struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

func (a *App) Chat(w http.ResponseWriter, r *http.Request) {
	urlpath := r.URL.Path[len("/chat/"):]
	switch r.Method {
	case http.MethodGet:
		templates.Render(w, "chat.html", map[string]interface{}{
			"Src": "/file/" + urlpath,
		})
	case http.MethodPost:
		log.Print("Post request to chat")

		var chathis []ChatHis
		if err := json.NewDecoder(r.Body).Decode(&chathis); err != nil {
			http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
			return
		}
		filepath := urlpath
		filename := urlpath
		strings.ReplaceAll(filename, "/", "_")

		responseChan, err := a.chatComplete(chathis, filepath, filename, r.Context())
		if err != nil {
			http.Error(w, "Error completing chat: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Cache-Control", "no-cache")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		for resp := range responseChan {
			fmt.Fprintf(w, "%s", resp)
			fmt.Print(resp)
			flusher.Flush()
		}

		flusher.Flush()
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type ChatRequest struct {
	Content  json.RawMessage `json:"content"`
	Prompt   string          `json:"message"`
	FilePath string          `json:"filePath"`
	FileName string          `json:"fileName"`
}

// chatComplete will complete the chat given the history
func (a *App) chatComplete(his []ChatHis, filepath, filename string, ctx context.Context) (chan string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(a.config.GEMINI_API_KEY))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	uri := ""
	if u, has := a.chatcache.Get(filepath); has {
		uri = u
		log.Print("Cache hit")
	} else {
		opts := genai.UploadFileOptions{DisplayName: filename}
		fileReader, err := a.helper.FileReader(filepath)
		if err != nil {
			return nil, err
		}
		doc, err := client.UploadFile(ctx, "", fileReader, &opts)
		if err != nil {
			return nil, err
		}
		uri = doc.URI
		a.chatcache.Set(filepath, uri, time.Hour*48) // Cache for 2 days
	}

	model := client.GenerativeModel("gemini-1.5-flash")

	cs := model.StartChat()
	cs.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.FileData{URI: uri},
				genai.Text("Please answer the following questions according to the content of the file."),
				genai.Text("Please answer the user's questions in Markdown format."),
			},
			Role: "user",
		},
	}

	for _, c := range his[:len(his)-1] {
		cs.History = append(cs.History, &genai.Content{
			Parts: []genai.Part{genai.Text(c.Text)},
			Role:  c.Role,
		})
	}

	responseChan := make(chan string)
	iter := cs.SendMessageStream(ctx, genai.Text(his[len(his)-1].Text))
	go func() {
		defer close(responseChan)
		for {
			res, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			for _, part := range res.Candidates[0].Content.Parts {
				responseChan <- fmt.Sprint(part)
			}
		}
	}()
	return responseChan, nil
}
