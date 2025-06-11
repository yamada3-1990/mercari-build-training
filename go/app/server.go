package app

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Server struct {
	// Port is the port number to listen on.
	Port string
	// ImageDirPath is the path to the directory storing images.
	ImageDirPath string
}

// Run is a method to start the server.
// This method returns 0 if the server started successfully, and 1 otherwise.
func (s Server) Run() int {
	// set up logger
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &opts))
	slog.SetDefault(logger)

	// set up CORS settings
	frontURL, found := os.LookupEnv("FRONT_URL")
	if !found {
		frontURL = "http://localhost:3000"
	}

	// STEP 5-1: set up the database connection
	db, err := sql.Open("sqlite3", "db/mercari.sqlite3")
	if err != nil {
		slog.Error("failed to open database: ", "error", err)
		return 1
	}
	defer db.Close()

	// set up handlers
	itemRepo, err := NewItemRepository(db)
	if err != nil {
		slog.Error("failed to create item repository: ", "error", err)
		return 1
	}
	h := &Handlers{imgDirPath: s.ImageDirPath, itemRepo: itemRepo}

	// set up routes
	// HTTPリクエストのルーティングを設定
	// handler:HTTPリクエストを処理する関数やメソッド
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.Hello)
	mux.HandleFunc("POST /items", h.AddItem)
	mux.HandleFunc("GET /items", h.GetItems)
	mux.HandleFunc("GET /images/{filename}", h.GetImage)
	mux.HandleFunc("GET /items/{item_id}", h.GetItemById)
	mux.HandleFunc("GET /search", h.SearchItemsByKeyword)

	// start the server
	slog.Info("http server started on", "port", s.Port)
	err = http.ListenAndServe(":"+s.Port, simpleCORSMiddleware(simpleLoggerMiddleware(mux), frontURL, []string{"GET", "HEAD", "POST", "OPTIONS"}))
	if err != nil {
		slog.Error("failed to start server: ", "error", err)
		return 1
	}

	return 0
}

type Handlers struct {
	// imgDirPath is the path to the directory storing images.
	imgDirPath string
	itemRepo   ItemRepository
}

type HelloResponse struct {
	Message string `json:"message"`
}

// Hello is a handler to return a Hello, world! message for GET / .
func (s *Handlers) Hello(w http.ResponseWriter, r *http.Request) {
	resp := HelloResponse{Message: "Hello, world!"}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetItems ハンドラーを実装 for GET /items
func (s *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	// GetAllメソッドを呼び出す
	items, err := s.itemRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Items []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
			Image    string `json:"image_name"`
		} `json:"items"`
	}{}

	for _, item := range items {
		response.Items = append(response.Items, struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
			Image    string `json:"image_name"`
		}{
			ID:       item.ID,
			Name:     item.Name,
			Category: item.Category,
			Image:    item.Image,
		})
	}

	// HTTPレスポンスのヘッダーを設定し、JSON形式でデータを書き込んでいます
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type AddItemRequest struct {
	Name     string `form:"name"`
	Category string `form:"category"`
	Image    []byte `form:"image"`
}

type AddItemResponse struct {
	Message string `json:"message"`
}

// parseAddItemRequest parses and validates the request to add an item.
func parseAddItemRequest(r *http.Request) (*AddItemRequest, error) {
	var req = &AddItemRequest{}

	// multipart/form-dataかを確認
	// リクエストがファイルアップロードを伴う multipart/form-data 形式であるかどうかを判断する
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err := r.ParseMultipartForm(32 << 20) // 32MBまで
		if err != nil {
			return nil, fmt.Errorf("failed to parse multipart form: %w", err)
		}

		req.Name = r.FormValue("name")
		req.Category = r.FormValue("category")

		// Get the image file
		file, header, err := r.FormFile("image")
		if err != nil {
			if !errors.Is(err, http.ErrMissingFile) {
				// ファイルがない場合はエラーにしない
				return nil, fmt.Errorf("failed to get image file: %w", err)
			}
			// ファイルがない場合は空のimageDataで続ける
		} else {
			defer file.Close()

			// jpgのみ受け付ける
			if !strings.HasSuffix(strings.ToLower(header.Filename), ".jpg") && !strings.HasSuffix(strings.ToLower(header.Filename), ".jpeg") {
				return nil, errors.New("only .jpg or .jpeg files are allowed")
			}

			// Read image data
			imageData, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read image data: %w", err)
			}
			if len(imageData) == 0 {
				return nil, errors.New("image data is empty")
			}

			req.Image = imageData
		}

	} else { // multipart/form-dataじゃなかったら
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}

		req.Name = r.FormValue("name")
		req.Category = r.FormValue("category")
	}

	// validaion
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Category == "" {
		return nil, errors.New("category is required")
	}

	return req, nil
}

// AddItem is a handler to add a new item for POST /items .
func (s *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseAddItemRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := "default.jpg"
	if len(req.Image) > 0 {
		fileName, err = s.storeImage(req.Image)
		if err != nil {
			slog.Error("failed to store image: ", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// デフォルト画像を読み込んで保存
		defaultImage, err := os.ReadFile(filepath.Join(s.imgDirPath, "default.jpg"))
		if err != nil {
			slog.Error("failed to read default image: ", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fileName, err = s.storeImage(defaultImage)
		if err != nil {
			slog.Error("failed to store default image: ", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	item := &Item{
		Name:     req.Name,
		Category: req.Category,
		Image:    strings.TrimPrefix(string(fileName), "images/"),
	}

	err = s.itemRepo.Insert(ctx, item)

	if err != nil {
		slog.Error("failed to store item: ", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf("item received: %s", item.Name)
	slog.Info(message)

	resp := AddItemResponse{Message: message}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// storeImage stores an image and returns the file path and an error if any.
// this method calculates the hash sum of the image as a file name to avoid the duplication of a same file
// and stores it in the image directory.
func (s *Handlers) storeImage(image []byte) (filePath string, err error) {
	// - calc hash sum
	hash := sha256.Sum256(image)
	// - build image file path
	// バックスラッシュをスラッシュに
	fileName := fmt.Sprintf("%x.jpg", hash)
	filePath = filepath.Join(s.imgDirPath, fileName)
	filePath = filepath.ToSlash(filePath)
	// - check if the image already exists
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}
	// - store image
	if err := os.WriteFile(filePath, image, 0644); err != nil {
		return "", fmt.Errorf("failed to write image file: %w", err)
	}
	// - return the image file path
	return filePath, nil
}

type GetImageRequest struct {
	FileName string // path value
}

// parseGetImageRequest parses and validates the request to get an image.
func parseGetImageRequest(r *http.Request) (*GetImageRequest, error) {
	req := &GetImageRequest{
		FileName: r.PathValue("filename"), // from path parameter
	} // validate the request
	if req.FileName == "" {
		return nil, errors.New("filename is required")
	}

	return req, nil
}

// GetImage is a handler to return an image for GET /images/{filename} .
// If the specified image is not found, it returns the default image.
func (s *Handlers) GetImage(w http.ResponseWriter, r *http.Request) {

	req, err := parseGetImageRequest(r)
	if err != nil {
		slog.Warn("failed to parse get image request: ", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	imgPath, err := s.buildImagePath(req.FileName)
	if err != nil {
		if !errors.Is(err, errImageNotFound) {
			slog.Warn("failed to build image path: ", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// when the image is not found, it returns the default image without an error.
		slog.Debug("image not found", "filename", imgPath)
		imgPath = filepath.Join(s.imgDirPath, "default.jpg")
	}

	slog.Info("returned image", "path", imgPath)
	http.ServeFile(w, r, imgPath)
}

// buildImagePath builds the image path and validates it.
// 画像を表示する際の処理
func (s *Handlers) buildImagePath(imageFileName string) (string, error) {
	imgPath := filepath.Join(s.imgDirPath, filepath.Clean(imageFileName))

	// to prevent directory traversal attacks
	// filepath.Rel(basepath, targetpath) は、basepath から targetpath への相対パスを返す
	rel, err := filepath.Rel(s.imgDirPath, imgPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid image path: %s", imgPath)
	}

	// validate the image suffix
	if !strings.HasSuffix(imgPath, ".jpg") && !strings.HasSuffix(imgPath, ".jpeg") {
		return "", fmt.Errorf("image path does not end with .jpg or .jpeg: %s", imgPath)
	}

	// check if the image exists
	// Stat: シンボリックリンクを辿って、リンク先のファイルやディレクトリの情報を返su
	_, err = os.Stat(imgPath)
	if err != nil {
		return imgPath, errImageNotFound
	}

	return imgPath, nil
}

/* GetItemById */
// リクエスト型をわざわざ宣言している理由: データの構造が明確, 
// リクエストに新しいパラメータを追加する場合、構造体にフィールドを追加するだけで済むなど
type GetItemByIdRequest struct {
	Id string
}

func parseGetItemByIdRequest(r *http.Request) (*GetItemByIdRequest, error) {
	req := &GetItemByIdRequest{
		Id: r.PathValue("item_id"),
	}

	// validate the request
	if req.Id == "" {
		return nil, errors.New("id is required")
	}

	return req, nil
}

func (s *Handlers) GetItemById(w http.ResponseWriter, r *http.Request) {
	req, err := parseGetItemByIdRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	item, err := s.itemRepo.GetItemById(r.Context(), req.Id)
	// エラーがerrItemNotFoundだったら404返す
	if err != nil {
		if errors.Is(err, errItemNotFound) {
			slog.Warn("item not exist: ", "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// jsonに変換
	jsonData, err := json.Marshal(item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

/* SearchItemsByKeyword */
type GetItemByKeywordRequest struct {
	Keyword string
}

func parseGetItemByKeywordRequest(r *http.Request) (*GetItemByKeywordRequest, error) {
	req := &GetItemByKeywordRequest{
		// クエリパラメータを取得
		Keyword: r.URL.Query().Get("keyword"),
	}

	// validation
	if req.Keyword == "" {
		return nil, errors.New("keyword is required")
	}

	return req, nil
}

func (s *Handlers) SearchItemsByKeyword(w http.ResponseWriter, r *http.Request) {
	req, err := parseGetItemByKeywordRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	items, err := s.itemRepo.SearchItemsByKeyword(r.Context(), req.Keyword)

	if err != nil {
		if errors.Is(err, errItemNotFound) {
			slog.Warn("item not exist: ", "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if items == nil {
		items = []Item{}
	}

	// jsonに変換
	jsonData, err := json.Marshal(items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
