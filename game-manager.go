package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	game "github.com/jiro4989/game-manager/game"

	keyboard "github.com/julienroland/keyboard-termbox"
	termbox "github.com/nsf/termbox-go"
	"github.com/urfave/cli"
)

const (
	coldef = termbox.ColorDefault
	colsel = termbox.ColorGreen

	idWidth             = 5
	gameNameWidth       = 30
	versionWidth        = 10
	gameFilePathWidth   = 40
	firstPlayWidth      = 13
	lastPlayWidth       = 13
	beforePlayTimeWidth = 16

	sep   = "| "
	max   = 2
	csvFn = "gameinfo.csv"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	app := cli.NewApp()
	app.Name = "game-manager"
	app.Usage = "ゲームのプレイ時間の記録を録ります。"
	app.Version = "1.0.0"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "edit,e",
			Usage: "アプリを追加します",
		},
	}

	app.Action = func(c *cli.Context) error {
		e := c.Bool("e")
		if e {
			var record [3]string
			fmt.Print("Enter app name >> ")
			fmt.Scan(&record[0])

			fmt.Print("Enter app version >> ")
			fmt.Scan(&record[1])

			fmt.Print("Enter app path >> ")
			fmt.Scan(&record[2])

			fmt.Println(record)

			return nil
		}

		// ゲーム情報ファイルの保存先を作成
		game.MkdirConfigDir(app.Name)
		// ゲームファイルを生成
		datas := game.ReadCsvFile(app.Name)

		if err := termbox.Init(); err != nil {
			log.Println(err)
			return err
		}
		defer termbox.Close()
		defer termbox.Clear(coldef, coldef)

		running := true
		selectedIndex := 1

		kb := keyboard.New()
		kb.Bind(func() { running = false }, "escape", "q")
		kb.Bind(func() {
			// カーソル移動 上
			selectedIndex--
			if selectedIndex < max {
				selectedIndex = max - 1
			}
		}, "up", "k")
		kb.Bind(func() {
			// カーソル移動 下
			selectedIndex++
			if len(datas)-1 < selectedIndex {
				selectedIndex = len(datas) - 1
			}
		}, "down", "j")
		kb.Bind(func() {
			// ゲーム起動とデータ更新
			sd := datas[selectedIndex]
			ep := sd[3]

			now := time.Now().Format("2006/01/02")

			// 最初のプレイ時間
			lp := strings.TrimSpace(sd[4])
			if lp == "" {
				datas[selectedIndex][4] = now
			}

			// 最後のプレイ時間
			datas[selectedIndex][5] = now

			dur := game.CalcTime(func() error {
				cmd := exec.Command(ep)
				err := cmd.Run()
				return err
			})
			sec := int(dur.Seconds())

			// 前回のプレイ時間
			datas[selectedIndex][6] = fmt.Sprintf("%d", sec)

			tt := strings.TrimSpace(sd[7])
			if tt == "" {
				tt = "0"
			}
			tti, err := strconv.Atoi(tt)
			if err != nil {
				log.Fatal(err)
			}

			// 合計プレイ時間
			datas[selectedIndex][7] = fmt.Sprintf("%d", tti+sec)

			// 更新した分のデータをファイルに保存
			err, err2 := game.SaveCsvFile(app.Name, &datas)
			if err != nil || err2 != nil {
				log.Fatal(err, err2)
			}
		}, "enter")

		for running {
			draw(datas, selectedIndex+max)
			kb.Poll(termbox.PollEvent())
		}

		return nil
	}

	app.Run(os.Args)
}

// 端末にデータ情報をテーブル上に描画する
func draw(datas [][]string, selectedIndex int) {
	termbox.Clear(coldef, coldef)

	drawLine(0, 0, "KeyInput >> q[uit], enter(play), j(down), k(up)", coldef, coldef)
	drawLine(0, 1, "=============================================================================================", coldef, coldef)

	for i, data := range datas {
		i += max
		var (
			fg = coldef
			bg = coldef
		)
		if i == selectedIndex {
			fg = colsel
		}

		x := 0
		drawLine(x, i, data[0], fg, bg)

		x += idWidth
		drawLine(x, i, sep+data[1], fg, bg)

		x += gameNameWidth
		drawLine(x, i, sep+data[2], fg, bg)

		x += versionWidth
		p := data[3]
		if gameFilePathWidth < len(p) {
			// 文字が収まりきらなかった時に...を表示する
			diff := len(p) - gameFilePathWidth + 6
			p = "..." + p[diff:]
		}
		drawLine(x, i, sep+p, fg, bg)

		x += gameFilePathWidth
		drawLine(x, i, sep+data[4], fg, bg)

		x += firstPlayWidth
		drawLine(x, i, sep+data[5], fg, bg)

		x += lastPlayWidth
		bt := strings.TrimSpace(data[6])
		if max < i {
			if bt != "" {
				// ヘッダは文字列だけれど、
				// 扱うデータは数値データなので
				// ヘッダ以外の処理の時だけ数値変換をする
				bti, err := strconv.Atoi(bt)
				if err != nil {
					log.Fatal(err)
				}
				bt = game.ParseTimeString(bti)
			}
		}
		drawLine(x, i, sep+bt, fg, bg)

		x += beforePlayTimeWidth
		tt := strings.TrimSpace(data[7])
		if max < i {
			if tt != "" {
				// ヘッダは文字列だけれど、
				// 扱うデータは数値データなので
				// ヘッダ以外の処理の時だけ数値変換をする
				tti, err := strconv.Atoi(tt)
				if err != nil {
					log.Fatal(err)
				}
				tt = game.ParseTimeString(tti)
			}
		}
		drawLine(x, i, sep+tt, fg, bg)
	}

	termbox.Flush()
}

// 一行のテキストを端末に書き込む
func drawLine(x, y int, text string, fg, bg termbox.Attribute) {
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		termbox.SetCell(x+i, y, runes[i], fg, bg)
	}
}
