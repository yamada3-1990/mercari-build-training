package app

import (
	"context"

	"database/sql"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")
var errItemNotFound = errors.New("item not found")

type Item struct {
	ID       int    `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Category string `json:"category"`
	Image    string `db:"image_name" json:"image_name"`
}

// item操作に関するメソッドを抽象化して定義している
// ItemRepository インターフェースは、Item 構造体 (または Item 構造体のスライス) を操作するメソッドをまとめている
// mockを利用してテストを行える
// 同じようなメソッドを持っているものをひとまとめにするための仕組み
// >「特定のメソッドの集合を定義し、そのメソッドを持つ型 (構造体など) は、そのインターフェースを実装しているとみなされる」
// interface:メソッドを使い回せる
// https://zenn.dev/logica0419/articles/understanding-go-interface
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]Item, error)
	GetItemById(ctx context.Context, item_id string) (Item, error)
	SearchItemsByKeyword(ctx context.Context, keyword string) ([]Item, error)
}

type itemRepository struct {
	db *sql.DB
}

// 返り値を増やした
// -> server.goのRun()でNewItemRepositoryのerrを検知できずに
// nilのitemRepoを使用したことによるnil参照panicを防ぐ
// NewItemRepositoryでデータベースの初期化に失敗した場合に、nilのitemRepoが使用されることを防ぐ
func NewItemRepository(db *sql.DB) (ItemRepository, error) {
	// items tableがなかったら作成
	q, err := os.ReadFile("db/items.sql")
	if err != nil {
		return &itemRepository{}, err
	}

	query := string(q)
	_, err = db.Exec(query)
	if err != nil {
		slog.Error("failed to create items table and categories table", "error", err)
		return nil, err
	}

	// データベース接続情報(db)を持つitemRepository構造体のインスタンスを作成し、そのポインタをItemRepositoryインターフェース型として返す。
	return &itemRepository{db: db}, nil
}

func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// Insert メソッドは、複数の関連するデータベース操作をまとめて実行する必要があるためトランザクションを使用
	// i.db は、itemRepository インスタンス i が保持しているデータベース接続
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// カテゴリが既に存在するか確認
	var categoryID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", item.Category).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			// カテゴリが存在しない場合は挿入
			_, err = tx.ExecContext(ctx, "INSERT INTO categories (name) VALUES (?)", item.Category)
			if err != nil {
				return err
			}
			// 挿入したカテゴリのIDを取得
			err = tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", item.Category).Scan(&categoryID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// itemsテーブルに挿入
	query := `INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)`
	_, err = tx.ExecContext(ctx, query, item.Name, categoryID, item.Image)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	// itemsとcategoriesをいったんinner join
	query := `
				SELECT
					items.id,
					items.name,
					categories.name AS category,
					items.image_name
				FROM
					items
				INNER JOIN
					categories ON items.category_id = categories.id;
			`

	// GetAll メソッドは単一のクエリで完結するため Query/Close を使用
	rows, err := i.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Item 構造体のスライス
	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// server.goのstoreImageで完結しているのでこっちのコードは使っていない
func StoreImage(fileName string, image []byte) error {
	savePath := filepath.Join("images", fileName)
	savePath = filepath.ToSlash(savePath)
	err := os.WriteFile(savePath, image, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (i *itemRepository) GetItemById(ctx context.Context, item_id string) (Item, error) {
	query := `
				SELECT 
					items.id, 
					items.name, 
					categories.name AS category, 
					items.image_name 
				FROM items
				INNER JOIN categories ON items.category_id = categories.id
				WHERE items.id = ?
			`
	row := i.db.QueryRow(query, item_id)
	var item Item
	// itemの各要素にセット
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
	if err != nil {
		if err == sql.ErrNoRows {
			return Item{}, errItemNotFound
		} else {
			return Item{}, err
		}
	}
	return item, nil
}

func (i *itemRepository) SearchItemsByKeyword(ctx context.Context, keyword string) ([]Item, error) {
	// itemsとcategoriesをいったんinner join
	query := `
				SELECT
								items.id,
								items.name,
								categories.name AS category,
								items.image_name
				FROM
								items
				INNER JOIN
								categories ON items.category_id = categories.id
				WHERE
								items.name LIKE ?
		`

	// queryの?部分がkeywordで置き換えられる
	// % はワイルドカード文字: 0文字以上の任意の文字列
	rows, err := i.db.Query(query, "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
