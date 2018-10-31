package types

import "golang.org/x/crypto/bcrypt"

// TODO: Default values
type Speciality struct {
	Id    int64  `json:"id"`
	Title string `sql:",unique" json:"title"`
}

type Department struct {
	Id    int64  `json:"id"`
	Title string `sql:",unique" json:"title"`
}

type Answer struct {
	Id          int64  `json:"answerId"`
	PageId      int64  `json:"-"`
	Text        string `json:"text"`
	Knowledge   int    `json:"knowledge"`
	Performance int    `json:"performance"`
	Sober       int    `json:"sober"`
	Prestige    int    `json:"prestige"`
	Connections int    `json:"connections"`
	Flags       string `json:"-"`
}

type Jumper struct {
	Id    int64  `json:'"jumpId"`
	Logic string `json:"-"`
}

// TODO: Flags as jsonb type (if possible)
// TODO: Array of answers for question pages

type Page struct {
	Id         int64  `json:"-"`
	NextPage   int64  `json:"-"`
	IsQuestion bool   `json:"-"`
	IsJumper   bool   `json:"-"`
	Year       int    `json:"year"`
	Dep        int64  `json:"-"`
	Spec       int64  `json:"-"`
	Text       string `json:"text"`
}

type User struct {
	Id          int64  `json:"-"`
	Login       string `sql:",unique" json:"login"`
	CurrentPage int64  `json:"-"`
	Password    string `json:"-"`
	Knowledge   int    `json:"knowledge"`
	Performance int    `json:"performance"`
	Sober       int    `json:"sober"`
	Prestige    int    `json:"prestige"`
	Connections int    `json:"connections"`
}

func (u User) Authenticate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

type CredentialsMessage struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
