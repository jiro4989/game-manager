package game

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	cfgnm = "gameinfo.csv"
	dcfg  = ".config"
	perm  = os.O_RDWR | os.O_CREATE
)

// int秒数を時間文字列(hh:mm:ss)に変換する
func ParseTimeString(sec int) string {
	h := sec / 60 / 60
	m := (sec - h*60*60) / 60
	s := sec - h*60*60 - m*60
	t := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	return t
}

// 引数で指定した関数の実行時間を計測して文字列として返却する
func CalcTime(cb func() error) (diffTime time.Duration) {
	// プレイ時間の計測
	startTime := time.Now()
	err := cb()
	if err != nil {
		log.Fatal(err)
	}
	endTime := time.Now()
	diffTime = endTime.Sub(startTime)
	return
}

// CSVファイルを読み込んで、二次元配列として返却する
func ReadCsvFile(appnm string) (lines [][]string) {
	home := getHomeEnv()
	cfp := filepath.Join(home, dcfg, appnm, cfgnm)
	data, err := ioutil.ReadFile(cfp)
	if err != nil {
		// ファイルが存在しない場合
		// 初期ファイルを生成して再度読み込み
		err, err2 := InitCsvFile(appnm)
		if err != nil || err2 != nil {
			log.Fatal(err)
		}

		data, err = ioutil.ReadFile(cfp)
		if err != nil {
			log.Fatal(err)
		}
	}

	dataStr := fmt.Sprintf("%s", data)
	r := csv.NewReader(strings.NewReader(dataStr))
	for {
		record, err := r.Read()
		if err == io.EOF {
			return
		}
		if err != nil {
			// 読み込みエラー
			log.Println(err)
			return
		}
		lines = append(lines, record)
	}
	return
}

// CSVファイルに書き込む
func SaveCsvFile(appnm string, datas *[][]string) (err, err2 error) {
	home := getHomeEnv()
	cfp := filepath.Join(home, dcfg, appnm, cfgnm)
	f, err := os.OpenFile(cfp, perm, 0666)
	defer f.Close()
	if err != nil {
		log.Print(err)
		return
	}
	w := csv.NewWriter(f)
	err2 = w.WriteAll(*datas)
	return
}

// 環境変数HOMEを取得するが、
// HOMEが空で且つ環境がWindowsならAPPDATAを使用する
func getHomeEnv() string {
	home := os.Getenv("HOME")
	if home == "" && runtime.GOOS == "windows" {
		// WindowsでHOME環境変数が定義されていない場合
		home = os.Getenv("APPDATA")
	}
	return home
}

// 設定ファイルの保存先ディレクトリを作成する
func MkdirConfigDir(appnm string) error {
	home := getHomeEnv()
	// 保存先ディレクトリ
	dir := filepath.Join(home, dcfg, appnm)
	err := os.MkdirAll(dir, os.ModeDir)
	return err
}

// ヘッダ情報のみのCSVファイルを作成する
func InitCsvFile(appnm string) (err, err2 error) {
	home := getHomeEnv()
	cfp := filepath.Join(home, dcfg, appnm, cfgnm)

	f, err := os.OpenFile(cfp, perm, 0666)
	if err != nil {
		return
	}
	defer f.Close()
	w := csv.NewWriter(f)

	data := [][]string{
		{
			"id",
			"game_name",
			"version",
			"game_file_patn",
			"reg_date",
			"last_play",
			"bef_play_time",
			"total",
		},
	}
	err2 = w.WriteAll(data)
	return
}
