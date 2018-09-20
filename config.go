package main

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"

	"gopkg.in/yaml.v2"
)

type User struct {
	Name        string `yaml:"name"`
	Password    string `yaml:"password"`
	ClusterRole string `yaml:"clusterRole"`
	Admin       bool   `yaml:"admin"`
}

func (u *User) isAdmin() bool {
	return u.Admin
}

type Users struct {
	Users []User `yaml:"users"`
}

type Config struct {
	Users       map[string]User
	GinAccounts *gin.Accounts
	Port        int
}

func (c *Config) Load(path string) *Config {
	if c.Users == nil {
		c.Users = make(map[string]User)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log().Printf("yamlFile.Get err #%v ", err)
		return c.LoadDefaults()
	}

	users := &Users{}
	err = yaml.Unmarshal([]byte(data), users)
	if err != nil {
		log().Error(err)
		return c.LoadDefaults()
	}

	c.LoadUsers(users)
	return c
}

func (c *Config) LoadUsers(u *Users) {
	log().Warn("loading users")
	users := make(map[string]User)
	accounts := make(gin.Accounts)

	for _, user := range u.Users {
		users[user.Name] = user
		accounts[user.Name] = user.Password
	}

	c.Users = users
	c.GinAccounts = &accounts

	return
}

func (c *Config) LoadDefaults() *Config {
	log().Warn("loading default config")
	defaults := []User{
		{Name: "admin", Password: "admin", ClusterRole: "admin", Admin: true},
		{Name: "edit", Password: "edit", ClusterRole: "edit", Admin: false},
		{Name: "view", Password: "view", ClusterRole: "view", Admin: false},
	}
	users := make(map[string]User)
	accounts := make(gin.Accounts)

	for _, user := range defaults {
		users[user.Name] = user
		accounts[user.Name] = user.Password
	}

	return c
}
