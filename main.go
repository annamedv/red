package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/rivo/tview"
	"github.com/satyrius/gonx"
)

var (
	duration    time.Duration
	distance    int
	format      string
	nginxConfig string
	nginxFormat string
	showHelp    bool
	keys        []string

	app   *tview.Application
	table *tview.Table
	store *Store
)

const (
	trendColumn int = iota
	countColumn
	firstDataColumn
)

func init() {
	flag.DurationVar(&duration, "trend", 10*time.Second, "duration of trend")
	flag.IntVar(&distance, "distance", 3, "levenshtein distance for combining similar log entities")
	flag.StringVar(&format, "format", "json", "stdin format")
	flag.StringVar(&nginxConfig, "nginx-config", "/etc/nginx/nginx.conf", "nginx config file")
	flag.StringVar(&nginxFormat, "nginx-format", "main", "nginx log_format name")
	flag.BoolVar(&showHelp, "help", false, "show help")
}

func main() {
	flag.Parse()
	keys = flag.Args()

	if showHelp {
		flag.Usage()
		os.Exit(2)
	}

	store = NewStore(duration, distance, keys)
	app = tview.NewApplication()

	viewerOpen := false
	viewer := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	viewer.SetBorder(true)

	table = tview.NewTable().
		SetFixed(1, 2).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				table.SetSelectable(false, false)
			}
			if key == tcell.KeyEnter {
				table.SetSelectable(true, false)
			}
		})
	renderColumns()

	flex := tview.NewFlex()
	flex.AddItem(table, 0, 1, true)
	app.SetRoot(flex, true)

	showRowData := func() {
		store.RLock()
		row, _ := table.GetSelection()
		if row == 0 {
			row = 1
		}
		data := store.Get(row - 1).GetData()
		store.RUnlock()

		text, err := prettyjson.Marshal(data)
		if err != nil {
			panic(err)
		}
		viewer.SetText(tview.TranslateANSI(string(text)))
		viewer.ScrollToBeginning()
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyUp {
			table.SetSelectable(true, false)
			if viewerOpen {
				showRowData()
			}
		}
		if event.Key() == tcell.KeyEnter && !viewerOpen {
			viewerOpen = true
			flex.AddItem(viewer, 0, 1, false)
			showRowData()
		}
		if event.Key() == tcell.KeyEsc && viewerOpen {
			viewerOpen = false
			flex.RemoveItem(viewer)
		}
		return event
	})

	switch format {
	case "json":
		go read()
	case "nginx":
		go readNginx()
	}

	go draw()
	go shift(duration)

	err := app.Run()
	if err != nil {
		panic(err)
	}
}

func renderColumns() {
	headerCell := func(s string) *tview.TableCell {
		return tview.NewTableCell(s).
			SetBackgroundColor(tcell.ColorRed).
			SetTextColor(tcell.ColorBlack).
			SetAlign(tview.AlignCenter).
			SetSelectable(false)
	}

	table.SetCell(0, trendColumn, headerCell("trend"))
	table.SetCell(0, countColumn, headerCell("count"))
	for i, key := range keys {
		table.SetCell(0, firstDataColumn+i, headerCell(key))
	}
}

func update(value map[string]interface{}) {
	store.Lock()
	if len(keys) == 0 {
		keys = mapKeys(value)
		store.SetKeys(keys)
		app.QueueUpdate(renderColumns)
	}
	store.Push(value)
	store.Unlock()
}

func read() {
	dec := json.NewDecoder(os.Stdin)
	for dec.More() {
		var value map[string]interface{}
		err := dec.Decode(&value)
		if err != nil {
			log.Println(err)
			app.Stop()
		}

		update(value)
	}
}

func readNginx() {
	config, err := os.Open(nginxConfig)
	if err != nil {
		panic(err)
	}
	defer config.Close()

	reader, err := gonx.NewNginxReader(os.Stdin, config, nginxFormat)
	if err != nil {
		panic(err)
	}
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		// Process the record... e.g.
		fmt.Printf("Parsed entry: %+v\n", rec)
	}
}

func shift(duration time.Duration) {
	for {
		store.Lock()
		store.Shift()
		store.Unlock()
		time.Sleep(duration / trendSize)
	}
}

func draw() {
	for {
		app.QueueUpdateDraw(func() {
			store.RLock()
			defer store.RUnlock()

			row := 1
			for ; row < table.GetRowCount(); row++ {
				data := store.Get(row - 1)
				table.GetCell(row, trendColumn).SetText(Spark(data.GetTrend()))
				table.GetCell(row, countColumn).SetText(data.GetCount())
				for j := 0; j < len(keys); j++ {
					text := fmt.Sprintf("%v", data.Get(keys[j]))
					table.GetCell(row, firstDataColumn+j).SetText(text)
				}
			}

			for ; row <= store.Len(); row++ {
				data := store.Get(row - 1)
				table.SetCell(row, trendColumn, tview.NewTableCell(Spark(data.GetTrend())).
					SetSelectable(false))
				table.SetCell(row, countColumn, tview.NewTableCell(data.GetCount()).
					SetSelectable(false))
				for j := 0; j < len(keys); j++ {
					text := fmt.Sprintf("%v", data.Get(keys[j]))
					table.SetCellSimple(row, firstDataColumn+j, text)
				}
			}
		})
		time.Sleep(100 * time.Millisecond)
	}
}
