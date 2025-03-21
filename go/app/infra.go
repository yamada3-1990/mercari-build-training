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

	return &itemRepository{db: db}, nil
}

func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
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
