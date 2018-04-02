package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type Configuration struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	ServerHost  string `json:"serverhost"`
	ServerPort  string `json:"serverport"`
	DBType      string `json:"dbtype"`
	DBUser      string `json:"dbuser"`
	DBPass      string `json:"dbpass"`
	DBName      string `json:"dbname"`
	DBHost      string `json:"dbhost"`
	DBPort      string `json:"dbport"`
	DBSSLMode   string `json:"dbsslmode"`
	BasePrefix  string `json:"baseprefix"`
	StaticDir   string `json:"staticdir"`
	TemplateDir string `json:"templatedir"`
	TagsDir     string `json:"tagsdir"`
	ExtendDir   string `json:"extenddir"`
	CookieKey   string `json:"cookiekey"`
}

func (c *Configuration) GetConnString() string {
	if c.DBType == "postgresql" {
		parameters := []string{}
		if c.DBUser != "" {
			parameters = append(parameters, fmt.Sprintf("user=%s", c.DBUser))
		}

		if c.DBPass != "" {
			parameters = append(parameters, fmt.Sprintf("password=%s", c.DBPass))
		}

		if c.DBName != "" {
			parameters = append(parameters, fmt.Sprintf("dbname=%s", c.DBName))
		}

		if c.DBHost != "" {
			parameters = append(parameters, fmt.Sprintf("host=%s", c.DBHost))
		}

		if c.DBPort != "" {
			parameters = append(parameters, fmt.Sprintf("port=%s", c.DBPort))
		}

		if c.DBSSLMode != "" {
			parameters = append(parameters, fmt.Sprintf("sslmode=%s", c.DBSSLMode))
		}

		return strings.Join(parameters, " ")
	} else if c.DBType == "sqlite3" {
		return c.DBName
	} else {
		return ""
	}
}

func Read(filename string) (Configuration, error) {
	config := Configuration{}

	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
