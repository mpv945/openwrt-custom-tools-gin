package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/db"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/model"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// GetByID 查询
func (r *UserRepository) GetByID(id int64) (*model.User, error) {

	var user model.User

	query := `
	SELECT id,name,email,created_at
	FROM users
	WHERE id = ?
	`

	err := db.DB.Get(&user, query, id)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByID(ctx context.Context, id int64) (*model.User, error) {

	query := `
	SELECT id, name, email
	FROM users
	WHERE id = ?
	`

	var user model.User

	// 超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := db.DB.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {

		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

// List 查询列表
func (r *UserRepository) List() ([]model.User, error) {

	var users []model.User

	query := `
	SELECT id,name,email,created_at
	FROM users
	ORDER BY id DESC
	`

	err := db.DB.Select(&users, query)

	return users, err
}

// Create 插入
func (r *UserRepository) Create(user *model.User) error {

	query := `
	INSERT INTO users(name,email)
	VALUES(?,?)
	`

	result, err := db.DB.Exec(query, user.Name, user.Email)

	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	// PostgreSQL Insert 推荐 RETURNING: INSERT ... RETURNING id 比 LastInsertId() 更好。
	user.ID = id

	return nil
}

// Update 更新
func (r *UserRepository) Update(user *model.User) error {

	query := `
	UPDATE users
	SET name=?,email=?
	WHERE id=?
	`

	_, err := db.DB.Exec(query,
		user.Name,
		user.Email,
		user.ID,
	)
	// 事务
	/*tx, err := db.DB.Beginx()

	_, err = tx.Exec("UPDATE account SET balance=balance-100 WHERE id=?", 1)

	_, err = tx.Exec("UPDATE account SET balance=balance+100 WHERE id=?", 2)

	tx.Commit()*/

	return err
}

// Delete 删除
func (r *UserRepository) Delete(id int64) error {

	query := `DELETE FROM users WHERE id=?`

	_, err := db.DB.Exec(query, id)

	return err
}

// PostgreSQL 占位符是
/*
type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {

	query := `
	INSERT INTO users(name,email)
	VALUES($1,$2)
	RETURNING id
	`

	return db.DB.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
	).Scan(&user.ID)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {

	query := `
	SELECT id,name,email,created_at
	FROM users
	WHERE id=$1
	`

	var user model.User

	err := db.DB.GetContext(ctx, &user, query, id)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {

	query := `
	SELECT id,name,email,created_at
	FROM users
	ORDER BY id DESC
	`

	var users []model.User

	err := db.DB.SelectContext(ctx, &users, query)

	return users, err
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {

	query := `
	UPDATE users
	SET name=$1,email=$2
	WHERE id=$3
	`

	_, err := db.DB.ExecContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.ID,
	)

	return err
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {

	query := `DELETE FROM users WHERE id=$1`

	_, err := db.DB.ExecContext(ctx, query, id)

	return err
}
*/
