package main

import (
	// "encoding/json"
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anacrolix/torrent"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

const sintel = "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent"

const quantum = time.Second / 3

func getMagnet() string {
	scanner := bufio.NewScanner(os.Stdin)
	// reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter a magnet link: ")
	scanner.Scan()
	magnet := scanner.Text()
	if len(magnet) == 0 {
		return sintel
	}
	return magnet
}

func prepTorrent(magnet string) (*torrent.Client, *torrent.Torrent) {
	c, _ := torrent.NewClient(nil)
	t, _ := c.AddMagnet(magnet)
	<-t.GotInfo()
	return c, t
}

func getTorrentParagraph(t *torrent.Torrent) *widgets.Paragraph {
	info := t.Info()
	out := widgets.NewParagraph()
	out.Title = "Torrent Info"
	out.Text = fmt.Sprintf("Torrent name: %v\nLength: %d ", info.Name, info.Length)
	return out
}

func getTorrentFilesList(t *torrent.Torrent) *widgets.Table {
	// TODO: table headers
	// TODO: better filenames (show extensions)
	out := widgets.NewTable()
	out.Title = "Files"
	out.RowSeparator = false
	out.TextAlignment = ui.AlignRight
	files := t.Info().Files
	out.Rows = make([][]string, len(files))
	for i, fi := range files {
		out.Rows[i] = []string{
			fmt.Sprintf("%v ", fi.Path),
			fmt.Sprintf("%d ", fi.Length),
		}
	}

	out.ColumnResizer = func() {
		width := out.Inner.Dx()
		maxNumsWidth := 0
		for _, row := range out.Rows {
			numLen := len(row[1])
			if numLen > maxNumsWidth {
				maxNumsWidth = numLen
			}
		}
		out.ColumnWidths = []int{
			width - maxNumsWidth - 1, // Make space for the divider.
			maxNumsWidth,
		}
	}
	return out
}

func main() {
	var magnet = sintel
	if len(os.Args) < 2 {
		fmt.Println("No magnet provided; torrenting Sintel.")
	} else {
		magnet = os.Args[1]
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// TODO: load placeholder UI.
	cli, tor := prepTorrent(magnet)
	defer cli.Close()

	textRow := ui.NewRow(0.25, ui.NewCol(0.5, getTorrentParagraph(tor)), ui.NewCol(0.5, getTorrentFilesList(tor)))

	g := widgets.NewGauge()
	g.Percent = 0
	gaugeRow := ui.NewRow(0.25, g)

	p1 := widgets.NewPlot()
	p1.Title = "Download Speed"
	p1.Marker = widgets.MarkerDot
	p1.Data = [][]float64{
		[]float64{0}, // Last observed download speed datum.
		[]float64{0}, // Average download speed from last period.
	}
	p1.AxesColor = ui.ColorWhite
	p1.LineColors[0] = ui.ColorYellow
	p1.DrawDirection = widgets.DrawRight
	statsRow := ui.NewRow(0.5, p1)

	pd := &progressDelta{
		fromMoment:   time.Now(),
		fromProgress: int64(0),
	}

	updateGauge := func() {
		g.Percent, g.Label = getStatsString(tor)
	}

	updatePlot := func() {
		read := tor.Stats().BytesReadUsefulData
		read64 := read.Int64()
		appendToPlot(p1, 0, pd.getUpdate(read64))
		appendToPlot(p1, 1, func(ns []float64) float64 {
			s := float64(0)
			for _, n := range ns {
				s += n
			}
			return float64(s) / float64(len(ns))
		}(p1.Data[0]))
		// Update title with ETA.
		p1.Title = fmt.Sprintf("Download Speed • ETA: %v", func() string {
			toRead := tor.Length() - read64
			lastRate := int64(p1.Data[1][len(p1.Data[1])-1])
			if lastRate == 0 {
				return "∞"
			}
			quantaRemaining := toRead / lastRate
			return time.Duration(quantum.Nanoseconds() * quantaRemaining).Round(time.Second).String()
		}())
	}

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(textRow, gaugeRow, statsRow)
	ui.Render(grid)

	uiEvents := ui.PollEvents()
	// FIXME: use a custom wrapper for the torrent for stats updates.
	ticker := time.NewTicker(quantum).C
	tor.DownloadAll()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>": // quit
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
			case "g":
				g.Percent++
			}
		case <-ticker:
			updateGauge()
			updatePlot()
		}
		ui.Render(grid)
	}
}

func getStatsString(t *torrent.Torrent) (int, string) {
	read := t.Stats().BytesReadUsefulData
	read64 := read.Int64()
	floatPercentage := float64(100) * (float64(read64) / float64(t.Length()))
	s := fmt.Sprintf("Downloaded %d/%d: %f%%", read.Int64(), t.Length(), floatPercentage)
	return int(floatPercentage), s
}

type progressDelta struct {
	fromMoment   time.Time
	fromProgress int64
}

func (pd *progressDelta) getUpdate(newProgress int64) float64 {
	// Return bytes per second for last period.
	lastFromMoment := pd.fromMoment
	pd.fromMoment = time.Now()
	elapsedSeconds := time.Since(lastFromMoment).Seconds()
	lastFromProgress := pd.fromProgress
	pd.fromProgress = newProgress
	return float64(newProgress-lastFromProgress) / elapsedSeconds
}

func appendToPlot(plt *widgets.Plot, si int, datum float64) {
	plt.Data[si] = append(plt.Data[si], datum)
	newLen := len(plt.Data[si])
	if len(plt.Data[si]) > plt.Inner.Dx()-5 {
		plt.Data[si] = plt.Data[si][newLen-(plt.Inner.Dx()-5):]
	}
}
