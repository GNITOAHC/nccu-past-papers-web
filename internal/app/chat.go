package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func (a *App) Chat(w http.ResponseWriter, r *http.Request) {
	urlpath := r.URL.Path[len("/chat/"):]
	a.tmplExecute(w, []string{"templates/chat.html"}, map[string]interface{}{
		"Src": "/file/" + urlpath,
	})
}

type ChatRequest struct {
	Content  json.RawMessage `json:"content"`
	Prompt   string          `json:"message"`
	FilePath string          `json:"filePath"`
	FileName string          `json:"fileName"`
}

func (a *App) ChatEndpoint(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}

	type ChatHis struct {
		Role string `json:"role"`
		Text string `json:"text"`
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
