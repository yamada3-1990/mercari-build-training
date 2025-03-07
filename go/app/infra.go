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
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `json:"category"`
	Image    string `db:"image" json:"image"`
}

type Category struct {
	Name string `db:"category" json:"category"`
}

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
func NewItemRepository(db *sql.DB) (ItemRepository, error) {
	// items tableがなかったら作成
	query := `
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            category_id INTEGER,
            image_name TEXT NOT NULL,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        );
    `
	_, err := db.Exec(query)
	if err != nil {
		slog.Error("failed to create items table", "error", err)
		return nil, err
	}

	// categories tableが無かったら作成
	query = `
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE
        );
    `
	_, err = db.Exec(query)
	if err != nil {
		slog.Error("failed to create categories table: ", "error", err)
		return nil, err
	}
	return &itemRepository{db: db}, nil
}

func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	var categoryID int

	// categories tableから(categories tableの)name = item.Categoryのidを取得
	err := i.db.QueryRow("SELECT id FROM categories WHERE name = ?", item.Category).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			// 該当する行がなかったら = 新しいカテゴリーだったら
			// categories tableのnameにitem.Categoryの値をinsert
			res, err := i.db.Exec("INSERT INTO categories (name) VALUES (?)", item.Category)
			if err != nil {
				return err
			}
			// 最後に挿入された自動採番(AUTOINCREMENT)のidを取得
			id, err := res.LastInsertId()
			if err != nil {
				return err
			}
			categoryID = int(id)
		} else {
			return err
		}
	}
	// insert
	query := "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)"
	_, err = i.db.Exec(query, item.Name, categoryID, item.Image)
	if err != nil {
		return err
	}

	return nil
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
	rows, err := i.db.Query(query)
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
	query := "SELECT id, name, category_id, image_name FROM items WHERE id = ?"
	row := i.db.QueryRow(query, item_id)
	var item Item
	var categoryID int
	err := row.Scan(&item.ID, &item.Name, &categoryID, &item.Image)
	if err != nil {
		if err == sql.ErrNoRows {
			return Item{}, errItemNotFound
		} else {
			return Item{}, err
		}
	}
	//categoryIDからcategoryNameを取得
	err = i.db.QueryRow("SELECT name from categories where id = ?", categoryID).Scan(&item.Category)
	if err != nil {
		return Item{}, err
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
