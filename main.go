package main

import (
	// "encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/anacrolix/torrent"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

// const magnet = "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent"

const magnet = "magnet:?xt=urn:btih:50247341e9c86396f26468c80730a494768696e2&dn=There.Will.Be.Blood.2007.1080p.BluRay.x264.anoXmous&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Fzer0day.ch%3A1337&tr=udp%3A%2F%2Fopen.demonii.com%3A1337&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Fexodus.desync.com%3A6969" // THERE WILL BE BLOOD

const quantum = time.Second / 3

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
	out.Text = fmt.Sprintf("Torrent name: %v\nLength: %d\nFiles:", info.Name, info.Length)
	// TODO: list files in separate widget.
	for i, fi := range info.Files {
		out.Text += fmt.Sprintf("\n%d. %v [%d] ", i, fi.Path, fi.Length)
	}
	return out
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// TODO: load placeholder UI.
	// TODO: get user input for magnet link.

	cli, tor := prepTorrent(magnet)
	defer cli.Close()

	textRow := ui.NewRow(1.0/3, getTorrentParagraph(tor))

	g := widgets.NewGauge()
	g.Percent = 0
	gaugeRow := ui.NewRow(1.0/3, g)

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
	statsRow := ui.NewRow(1.0/3, p1)

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
