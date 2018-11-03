package types

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// TODO: Default values
// TODO: deny nulls
// TODO: Page string tags to simplify search and
// identification
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
	Knowledge   int    `json:"-"`
	Performance int    `json:"-"`
	Sober       int    `json:"-"`
	Prestige    int    `json:"-"`
	Connections int    `json:"-"`
	Flags       string `json:"-"`
}

type Page struct {
	Id          int64  `json:"-"`
	NextPage    int64  `json:"-"`
	IsQuestion  bool   `json:"-"`
	IsJumper    bool   `json:"-"`
	Year        int    `json:"year"`
	Dep         int64  `json:"-"`
	Spec        int64  `json:"-"`
	Text        string `json:"text"`
	JumperLogic string `json:"-"`
}

type User struct {
	Id          int64  `json:"-"`
	Login       string `sql:",unique" json:"-"`
	CurrentPage int64  `json:"-"`
	Password    string `json:"-"`
	Knowledge   int    `json:"knowledge"`
	Performance int    `json:"performance"`
	Sober       int    `json:"sober"`
	Prestige    int    `json:"prestige"`
	Connections int    `json:"connections"`
	Flags       string `json:"-"`
}

func (u User) Authenticate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func contains(slice []string, str string) bool {
	for _, a := range slice {
		if a == str {
			return true
		}
	}
	return false
}

func (u *User) MergeFlags(flags string) {
	flagsArr := strings.Split(flags, " ")
	userFlags := strings.Split(u.Flags, " ")
	for _, fl := range flagsArr {
		if contains(userFlags, fl) == false {
			userFlags = append(userFlags, fl)
		}
	}
	u.Flags = strings.Join(userFlags, " ")
}

func (u User) IsFlagSet(flag string) bool {
	flagsArr := strings.Split(u.Flags, " ")
	return contains(flagsArr, flag)
}

type CredentialsMessage struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
