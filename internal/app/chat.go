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
		// case http.MethodPut:
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

		// for i := 0; i < 5; i++ {
		// 	// Send a chunk of data
		// 	fmt.Fprintf(w, "Message %d\n", i+1)
		// 	log.Print("data: Message ", i+1)
		//
		// 	// Flush the buffer so the client receives the chunk immediately
		// 	flusher.Flush()
		//
		// 	// Simulate some processing delay
		// 	time.Sleep(1 * time.Second)
		// }
		flusher.Flush()
		// w.Write([]byte("Hello"))
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

	go func() {
		defer close(responseChan)

		resp, err := cs.SendMessage(ctx, genai.Text(his[len(his)-1].Text))
		if err != nil {
			log.Print("Error sending message:", err)
			return
		}

		// Process streaming response
		if resp.Candidates[0].Content != nil {
			for _, part := range resp.Candidates[0].Content.Parts {
				// Send each part to the channel
				responseChan <- fmt.Sprint(part)
			}
		}
	}()

	return responseChan, nil
}

func (a *App) ChatEndpoint(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}

	var chatHis []ChatHis
	if err := json.Unmarshal(req.Content, &chatHis); err != nil {
		http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	client, err := genai.NewClient(ctx, option.WithAPIKey(a.config.GEMINI_API_KEY))
	if err != nil {
		http.Error(w, "Error creating genai client: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	uri := ""
	if u, has := a.chatcache.Get(req.FilePath); has {
		uri = u
		log.Print("Cache hit")
	} else {
		opts := genai.UploadFileOptions{DisplayName: req.FileName}
		fileReader, err := a.helper.FileReader("ComputerScience/Algo_%E6%B2%88%E9%8C%B3%E5%9D%A4/111_2_Mid.pdf")
		if err != nil {
			http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		doc, err := client.UploadFile(r.Context(), "", fileReader, &opts)
		if err != nil {
			http.Error(w, "Error uploading file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		uri = doc.URI
		a.chatcache.Set(req.FilePath, uri, time.Hour*48) // Cache for 2 days
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	// model.SystemInstruction = &genai.Content{
	// 	Parts: []genai.Part{
	// 		genai.Text("Please answer the following questions according to the content of the file."),
	// 	},
	// }
	cs := model.StartChat()
	cs.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.FileData{URI: uri},
				genai.Text("Please answer the following questions according to the content of the file."),
			},
			Role: "user",
		},
	}

	for _, c := range chatHis {
		cs.History = append(cs.History, &genai.Content{
			Parts: []genai.Part{genai.Text(c.Text)},
			Role:  c.Role,
		})
	}

	resp, err := cs.SendMessage(r.Context(), genai.Text(req.Prompt))
	if err != nil {
		// log.Print("Error generating content: ", err)
		http.Error(w, "Error generating content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resMes := ""
	if resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			resMes += fmt.Sprint(part)
		}
	}

	// log.Print("Response: ", resMes)
	w.Write([]byte(resMes))
}
