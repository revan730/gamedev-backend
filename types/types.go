package types

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

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
	Text        string `json:"text" sql:"default:''"`
	Knowledge   int    `json:"-" sql:"default:0"`
	Performance int    `json:"-" sql:"default:0"`
	Sober       int    `json:"-" sql:"default:0"`
	Prestige    int    `json:"-" sql:"default:0"`
	Connections int    `json:"-" sql:"default:0"`
	Flags       string `json:"-" sql:"default:''"`
}

type Page struct {
	Id          int64  `json:"-"`
	NextPage    int64  `json:"-"`
	IsQuestion  bool   `json:"-"`
	IsJumper    bool   `json:"-"`
	Year        int    `json:"year" sql:"default:0"`
	Dep         int64  `json:"-" sql:"default:0"`
	Spec        int64  `json:"-" sql:"default:0"`
	Text        string `json:"text"`
	JumperLogic string `json:"-" sql:"default:''"`
}

type User struct {
	Id          int64  `json:"-"`
	Login       string `sql:",unique" json:"-"`
	CurrentPage int64  `json:"-" sql:"default:1"`
	Password    string `json:"-"`
	Knowledge   int    `json:"knowledge" sql:"default:0"`
	Performance int    `json:"performance" sql:"default:0"`
	Sober       int    `json:"soberness" sql:"default:0"`
	Prestige    int    `json:"prestige" sql:"default:0"`
	Connections int    `json:"connections" sql:"default:0"`
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

// Reset resets user's stats and current page to beggining
func (u *User) Reset() {
	u.Connections = 0
	u.Sober = 0
	u.Performance = 0
	u.Prestige = 0
	u.Knowledge = 0
	u.CurrentPage = 1
	u.Flags = ""
}

type CredentialsMessage struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
