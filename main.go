package main

import (
	"log"
	// "math"
	"encoding/json"

	"github.com/anacrolix/torrent"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

const magnet = "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent"

func getTorrentParagraph(magnet string) *widgets.Paragraph {
	c, _ := torrent.NewClient(nil)
	defer c.Close()
	t, _ := c.AddMagnet(magnet)
	<-t.GotInfo()
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

	textRow := ui.NewRow(1.0/3, getTorrentParagraph(magnet))

	g := widgets.NewGauge()
	g.Percent = 40
	g.Label = "ProgresS"
	gaugeRow := ui.NewRow(1.0/3, g)

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(textRow, gaugeRow)

	ui.Render(grid)
	// ui.Render(g)

	for e := range ui.PollEvents() {
		switch e.ID {
		case "<Resize>":
			payload := e.Payload.(ui.Resize)
			grid.SetRect(0, 0, payload.Width, payload.Height)
			ui.Clear()
			ui.Render(grid)
		case "q", "<C-c>":
			return
		case "g":
			g.Percent++
		}
		ui.Render(grid)
	}
}
