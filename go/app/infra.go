package app

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	// STEP 5-1: uncomment this line
	_ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")
var errItemNotFound = errors.New("item not found")

var Db *sql.DB

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image    string `db:"image" json:"image"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]Item, error)
	GetItemById(ctx context.Context, item_id string) (Item, error)
	SearchItemsByKeyword(ctx context.Context, keyword string) ([]Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// fileName is the path to the JSON file storing items.
	db *sql.DB
}

// NewItemRepository creates a new itemRepository.
// main.goを実行するディレクトリによってfileNameを変更する
func NewItemRepository(db *sql.DB) ItemRepository {
	return &itemRepository{db: db}
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// mercari.sqlite3に接続
	Db, _ := sql.Open("sqlite3", "db/mercari.sqlite3")
	defer Db.Close()

	// DBにitemをインサート
	query := "INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)"
	_, err := Db.Exec(query, item.Name, item.Category, item.Image)
	if err != nil {
		return err
	}

	return nil
}

// GetAll()
func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	// mercari.sqlite3に接続
	Db, _ := sql.Open("sqlite3", "db/mercari.sqlite3")
	defer Db.Close()

	query := "SELECT * FROM items"
	rows, _ := Db.Query(query)
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var i Item
		err := rows.Scan(&i.ID, &i.Name, &i.Category, &i.Image)
		if err != nil {
			return []Item{}, err
		}
		items = append(items, i)
	}

	return items, nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	// 保存先
	savePath := filepath.Join("images", fileName)

	// バックスラッシュをスラッシュに
	savePath = filepath.ToSlash(savePath)
	// ファイルを保存
	err := os.WriteFile(savePath, image, 0644)
	if err != nil {
		return err
	}

	return nil

}

// GetItemById()
func (i *itemRepository) GetItemById(ctx context.Context, item_id string) (Item, error) {
	// mercari.sqlite3に接続
	Db, _ := sql.Open("sqlite3", "db/mercari.sqlite3")
	defer Db.Close()

	query := "SELECT * FROM where id = ?"
	row := Db.QueryRow(query, item_id)
	var item Item
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
	if err != nil {
		if err == sql.ErrNoRows {
			return Item{}, errors.New("no row")
		} else {
			return Item{}, err
		}
	}

	return item, nil
}

// SearchItemsByKeyword()
func (i *itemRepository) SearchItemsByKeyword(ctx context.Context, keyword string) ([]Item, error) {
	// mercari.sqlite3に接続
	Db, _ := sql.Open("sqlite3", "db/mercari.sqlite3")
	defer Db.Close()

	query := "SELECT * FROM items WHERE name LIKE ?"
	rows, err := Db.Query(query, "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var i Item
		err := rows.Scan(&i.ID, &i.Name, &i.Category, &i.Image)
		if err != nil {
			return []Item{}, err
		}
		items = append(items, i)
	}

	return items, nil
}

/*

*** STEP 4 ***
GETとPOSTのリクエストの違いについて調べてみましょう
	->GET:  サーバーにリクエストを送信、リソースを取得
	->POST: サーバーにデータを送信、リソースの更新など

ブラウザで http://127.0.0.1:9000/items にアクセスしても {"message": "item received: <name>"} が返ってこないのはなぜでしょうか？
	-> server.go の route に GET /items がないから？

アクセスしたときに返ってくるHTTPステータスコードはいくつですか？
	-> 200 OK

それはどんな意味をもつステータスコードですか？
	-> リクエストが正常に処理された

ハッシュ化とはなにか？
	-> 特定のルール(ハッシュ関数)に基づいて値を変換すること

SHA-256 以外にどんなハッシュ関数があるか調べてみましょう
	-> SHA-3, MD5など >アルゴリズムの設計、セキュリティ強度、速度、用途が違う らしい

Log levelとは？
	-> ソフトウェアが記録するログ(どんな動作が行われたかの記録)の詳細度と重要度を調整するための仕組み

webサーバーでは、本番はどのログレベルまで表示する？
	-> INFO以上が一般的 開発環境だとDEBUG

port (ポート番号)
	-> コンピュータが通信に使用するプログラムを識別するための番号 HTTP:80 etc.

localhost, 127.0.0.1
	-> localhost: コンピューター自身を指し示すためのホスト名
	-> 127.0.0.1: IPv4における特別なIPアドレス

HTTPリクエストメソッド
	-> Webサーバーにどのような処理をするかを伝える役割
	-> GET/POST/PUT(更新)/PATCH(一部更新)/DELETE(削除)

HTTPステータスコード (1XX, 2XX, 3XX, 4XX, 5XXはそれぞれどんな意味を持ちますか？)
	-> 1XX: リクエストが受け付けられて処理が続いている(Informational)
	-> 2XX: リクエストが正常に完了(Success)
	-> 3XX: リクエストを完了するために追加のアクションが必要(Redirection)
	-> 4XX: リクエストに問題あり(Client Error)
	-> 5XX: サーバーがリクエストを処理できなかった(Server Error)


*** STEP 5 ***
jsonファイルではなくデータベース(SQLite)にデータを保存する利点は何がありますか？
	-> dbだとデータの整合性がとりやすい、データ操作・検索が効率的(jsonだとファイル全体を読む込む必要がある)

*/

// curlじゃなくて curl.exe で実行
// cd go してから go run cmd/api/main.go でサーバーを起動するなら
// main.go の実行ディレクトリは go/
