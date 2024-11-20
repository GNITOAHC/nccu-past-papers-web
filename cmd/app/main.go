package main

import (
	"log"
	"net/http"
	"past-papers-web/internal/app"
	"past-papers-web/internal/config"
	"past-papers-web/internal/helper"
)

func main() {
	app.StartServer()

	// 從配置文件加載設置
	cfg := config.NewConfig(".env") // 加載環境變量配置

	// 初始化 GitHub Tree 數據
	treeData := helper.InitTree(cfg)   // 初始化樹數據
	root := helper.ParseTree(treeData) // 解析樹結構

	// 註冊搜尋處理器
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		helper.SearchHandler(w, r, root)
	})

	// 啟動 HTTP 服務
	log.Println("Server is running on port 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
