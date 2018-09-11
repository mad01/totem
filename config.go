package main

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"

	"gopkg.in/yaml.v2"
)

type User struct {
	Name               string `yaml:"name"`
	Password           string `yaml:"password"`
	AccessLevel        string `yaml:"accessLevel"` // admin/edit/view
	ClusterRoleBinding string `yaml:"clusterRoleBinding"`
}

type Config struct {
	Users       []User `yaml:"users"`
	GinAccounts *gin.Accounts
	Port        int
}

func (c *Config) Load(path string) *Config {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log().Printf("yamlFile.Get err #%v ", err)
		return c.LoadDefaults()
	}

	err = yaml.Unmarshal([]byte(data), c)
	if err != nil {
		log().Error(err)
		return c.LoadDefaults()
	}

	c.GinAccounts = c.LoadGinAccounts()
	log().Info(c)
	return c
}

func (c *Config) LoadGinAccounts() *gin.Accounts {
	accounts := make(gin.Accounts)
	for _, user := range c.Users {
		accounts[user.Name] = user.Password
	}
	return &accounts
}

func (c *Config) LoadDefaults() *Config {
	log().Warn("loading default config")
	users := []User{
		{Name: "admin", Password: "admin", AccessLevel: "admin", ClusterRoleBinding: "admin"},
		{Name: "edit", Password: "edit", AccessLevel: "edit", ClusterRoleBinding: "edit"},
		{Name: "view", Password: "view", AccessLevel: "view", ClusterRoleBinding: "view"},
	}
	c.Users = users
	c.GinAccounts = c.LoadGinAccounts()
	return c
}
