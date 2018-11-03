package db

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/revan730/gamedev-backend/types"
	"golang.org/x/crypto/bcrypt"
)

type DatabaseClient struct {
	pg *pg.DB
}

func NewDBClient(addr, db, user, pass string) *DatabaseClient {
	DBClient := &DatabaseClient{}
	pgdb := pg.Connect(&pg.Options{
		User:         user,
		Addr:         addr,
		Password:     pass,
		Database:     db,
		MinIdleConns: 2,
	})
	DBClient.pg = pgdb
	return DBClient
}

func (d *DatabaseClient) Close() {
	d.pg.Close()
}

// CreateSchema creates database tables if they not exist
func (d *DatabaseClient) CreateSchema() error {
	for _, model := range []interface{}{(*types.User)(nil),
		(*types.Page)(nil),
		(*types.Answer)(nil),
		(*types.Department)(nil),
		(*types.Speciality)(nil)} {
		err := d.pg.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func (d *DatabaseClient) CreateUser(login, pass string) error {
	hash, err := HashPassword(pass)
	if err != nil {
		return err
	}
	user := &types.User{
		Login:    login,
		Password: hash,
	}

	return d.pg.Insert(user)
}

func (d *DatabaseClient) SaveUser(user *types.User) error {
	return d.pg.Update(user)
}

func (d *DatabaseClient) FindUser(login string) (*types.User, error) {
	user := &types.User{
		Login: login,
	}

	err := d.pg.Model(user).
		Where("login = ?", login).
		Select()
	if err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func (d *DatabaseClient) FindUserById(userId int64) (*types.User, error) {
	user := &types.User{
		Id: userId,
	}

	err := d.pg.Select(user)
	if err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func (d *DatabaseClient) FindPageById(pageId int64) (*types.Page, error) {
	page := &types.Page{
		Id: pageId,
	}
	err := d.pg.Select(page)
	if err != nil {
		return nil, err
	} else {
		return page, nil
	}
}

func (d *DatabaseClient) FindAnswerById(answerId int64) (*types.Answer, error) {
	answer := &types.Answer{
		Id: answerId,
	}
	err := d.pg.Select(answer)
	if err != nil {
		return nil, err
	} else {
		return answer, nil
	}
}

func (d *DatabaseClient) FindPageAnswers(pageId int64) ([]types.Answer, error) {
	var answers []types.Answer
	_, err := d.pg.Query(&answers, "SELECT * FROM answers WHERE page_id = ?", pageId)
	return answers, err
}
