package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/anacrolix/torrent"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

const magnet = "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent"

func prepTorrent(magnet string) (*torrent.Client, *torrent.Torrent) {
	c, _ := torrent.NewClient(nil)
	// NOTE: remember to defer c.Close()
	t, _ := c.AddMagnet(magnet)
	<-t.GotInfo()
	return c, t
}

func getTorrentParagraph(t *torrent.Torrent) *widgets.Paragraph {
	out := widgets.NewParagraph()
	r, _ := json.MarshalIndent(t.Info(), "", "  ")
	out.Text = string(r)
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
	// TODO: show list of torrent files.

	g := widgets.NewGauge()
	g.Percent = 0
	gaugeRow := ui.NewRow(1.0/3, g)

	stats := widgets.NewParagraph()
	_, stats.Text = getStatsString(tor)
	statsRow := ui.NewRow(1.0/3, stats)

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(textRow, gaugeRow, statsRow)

	ui.Render(grid)
	// ui.Render(g)

	tor.DownloadAll()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C // FIXME: use a custom wrapper for the torrent for stats updates.

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
			// Update the bar chart with progress.
			g.Percent, stats.Text = getStatsString(tor)
			g.Label = stats.Text
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
