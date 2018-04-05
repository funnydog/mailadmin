package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ServerHost   string `json:"serverhost"`
	ServerPort   string `json:"serverport"`
	ServerCert   string `json:"servercert"`
	ServerKey    string `json:"serverkey"`
	DBType       string `json:"dbtype"`
	DBUser       string `json:"dbuser"`
	DBPass       string `json:"dbpass"`
	DBName       string `json:"dbname"`
	DBHost       string `json:"dbhost"`
	DBPort       string `json:"dbport"`
	DBSSLMode    string `json:"dbsslmode"`
	BasePrefix   string `json:"baseprefix"`
	StaticPrefix string `json:"staticprefix"`
	StaticDir    string `json:"staticdir"`
	TemplateDir  string `json:"templatedir"`
	TagsDir      string `json:"tagsdir"`
	ExtendDir    string `json:"extenddir"`
	CookieKey    string `json:"cookiekey"`
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
