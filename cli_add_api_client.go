package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/learnin/go-multilog"
	"github.com/mattn/go-colorable"

	"github.com/learnin/batch-rest-controller/helpers"
	"github.com/learnin/batch-rest-controller/models"
)

const LOG_DIR = "log"
const LOG_FILE = LOG_DIR + "/cli_add_api_client.log"
const SALT = "jOArue9da9wfywrw89*(Yaqipkdoeojapiefhqoy*Oo"

var log *multilog.MultiLogger

func main() {
	if fi, err := os.Stat(LOG_DIR); os.IsNotExist(err) {
		if err := os.MkdirAll(LOG_DIR, 0755); err != nil {
			panic(err)
		}
	} else {
		if !fi.IsDir() {
			panic("ログディレクトリ " + LOG_DIR + " はディレクトリではありません。")
		}
	}
	logf, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer logf.Close()
	stdOutLogrus := logrus.New()
	stdOutLogrus.Out = colorable.NewColorableStdout()
	fileLogrus := logrus.New()
	fileLogrus.Out = logf
	fileLogrus.Formatter = &logrus.TextFormatter{DisableColors: true}
	log = multilog.New(stdOutLogrus, fileLogrus)

	app := cli.NewApp()
	app.Name = "cli-add-api-client"
	app.Version = "0.0.1"
	app.Author = "Manabu Inoue"
	app.Email = ""
	app.HideVersion = true
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "verbose mode. a lot more information output",
		},
		cli.BoolFlag{
			Name:  "version, V",
			Usage: "print the version",
		},
	}
	app.Usage = "add API client."
	app.Action = func(c *cli.Context) {
		log.Info("APIクライアントの登録を開始します。")
		defer log.Info("APIクライアントの登録を終了しました。")

		action(c)
	}
	app.Run(os.Args)
}

func hash(s string, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(s + salt))
	return hex.EncodeToString(hash.Sum(nil))
}

func action(c *cli.Context) {
	if len(c.Args()) == 0 {
		log.Error("クライアント名を指定してください。")
		return
	}
	clientName := c.Args()[0]
	// FIXME バリデーション

	newKey := hash(clientName, SALT)
	apiKey := models.ApiKey{
		ClientName: clientName,
		ApiKey:     newKey,
	}

	isVerbose := c.Bool("verbose")

	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		log.Error("DB接続に失敗しました。" + err.Error())
		return
	}
	defer ds.Close()

	if isVerbose {
		ds.LogMode(true)
	}

	if err := ds.GetDB().Create(&apiKey).Error; err != nil {
		log.Error("DB登録に失敗しました。" + err.Error())
		return
	}
	println(newKey)

}
