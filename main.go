package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	alog "github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

/* CONSTS */

// A quantum is the duration between UI progress updates.
const updatesPerSecond = 3
const quantum = time.Second / updatesPerSecond

/* TORRENT UTILS */

// getMagnet prompts the user to input a magnet link and reads the user input if
// no magnet link is provided as a command-line argument.
func getMagnet() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter a magnet link: ")
	scanner.Scan()
	magnet := scanner.Text()
	return magnet
}

// prepTorrent creates a torrent client and synchronously fetches the torrent
// info.
func prepTorrent(magnet string) (*torrent.Client, *torrent.Torrent) {
	log.Println("Configuring torrent client...")
	cfg := torrent.NewDefaultClientConfig()
	cfg.Debug = false
	cfg.Logger = alog.Discard
	c, err := torrent.NewClient(cfg)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Processing magnet...")
	t, err := c.AddMagnet(magnet)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Fetching torrent info...")
	<-t.GotInfo()
	return c, t
}

/* UI CONTENT GENERATORS & UPDATERS */

// byteLengthToString returns a human-readable version of a byte length; it
// converts to kB, MB, GB.
func byteLengthToString(length int64) string {
	kB := float64(1000)
	MB := float64(math.Pow(float64(1000), float64(2)))
	GB := float64(math.Pow(float64(1000), float64(3)))
	flength := float64(length)
	if flength > 10*GB {
		return fmt.Sprintf("%.2f GB", flength/GB)
	} else if flength > 10*MB {
		return fmt.Sprintf("%.2f MB", flength/MB)
	} else if flength > 10*kB {
		return fmt.Sprintf("%.2f kB", flength/kB)
	}
	return fmt.Sprintf("%d B", length)
}

// A progressTracker is a torrent download state; it records the last-processed
// moment and the progress at that moment.
type progressTracker struct {
	fromMoment   time.Time
	fromProgress int64
}

// getSpeedUpdate processes progress since PD was last updated, updates PD, and
// returns the updated download speed.
func (pd *progressTracker) getSpeedUpdate(newProgress int64) float64 {
	// Return bytes per second for last period.
	lastFromMoment := pd.fromMoment
	pd.fromMoment = time.Now()
	elapsedSeconds := time.Since(lastFromMoment).Seconds()
	lastFromProgress := pd.fromProgress
	pd.fromProgress = newProgress
	return float64(newProgress-lastFromProgress) / elapsedSeconds
}

// getTorrentDescription generates the torrent info text box contents for the UI.
func getTorrentDescription(t *torrent.Torrent) *widgets.Paragraph {
	info := t.Info()
	out := widgets.NewParagraph()
	out.Title = "Torrent Info"
	out.Text = fmt.Sprintf(
		"Torrent name: %v\nLength: %s",
		info.Name,
		byteLengthToString(t.Length()),
	)
	return out
}

// getTorrentFilesList generates the torrent files table for the UI.
func getTorrentFilesList(t *torrent.Torrent) *widgets.Table {
	// TODO: table headers.
	out := widgets.NewTable()
	out.Title = "Files"
	out.RowSeparator = false
	out.TextAlignment = ui.AlignLeft
	files := t.Info().Files
	out.Rows = make([][]string, len(files))
	for i, fi := range files {
		out.Rows[i] = []string{
			fmt.Sprintf("%s ", strings.Join(fi.Path, "/")),
			fmt.Sprintf("%s ", byteLengthToString(fi.Length)),
		}
	}
	return out
}

// getProgressGaugeLabel generates the gauge label for a particular torrent
// state.
func getProgressGaugeLabel(t *torrent.Torrent) (int, string) {
	read := t.Stats().BytesReadUsefulData
	read64 := read.Int64()
	floatPercentage := float64(100) * (float64(read64) / float64(t.Length()))
	s := fmt.Sprintf("↓%s of %s: %.1f%%",
		byteLengthToString(read.Int64()),
		byteLengthToString(t.Length()),
		floatPercentage,
	)
	return int(floatPercentage), s
}

// rotateIntoPlot appends the datum to the SIth series in the plot, and rotates
// previous values out of the plot if the resulting series length is longer than
// the plot width.
func rotateIntoPlot(plt *widgets.Plot, si int, datum float64) {
	plt.Data[si] = append(plt.Data[si], datum)
	newLen := len(plt.Data[si])
	if len(plt.Data[si]) > plt.Inner.Dx()-5 {
		plt.Data[si] = plt.Data[si][newLen-(plt.Inner.Dx()-5):]
	}
}

/* INTERACTION AND LOOP */

func main() {
	var magnet string
	if len(os.Args) < 2 {
		magnet = getMagnet()
	} else {
		magnet = os.Args[1]
	}
	if len(magnet) == 0 {
		log.Println("No magnet provided. Torrenting Sintel (2010).")
		magnet = sintelMagnet
	}

	// TODO: load placeholder UI.
	cli, tor := prepTorrent(magnet)
	defer cli.Close()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Add torrent description and files list.
	textRow := ui.NewRow(0.25,
		ui.NewCol(0.5, getTorrentDescription(tor)),
		ui.NewCol(0.5, getTorrentFilesList(tor)),
	)

	// Add progress gauge.
	g := widgets.NewGauge()
	g.Percent = 0
	gaugeRow := ui.NewRow(0.25, g)
	updateGauge := func() { // TODO: refactor out updaters.
		g.Percent, g.Label = getProgressGaugeLabel(tor)
	}

	// TODO: make a custom progplot object with named references to its data that
	// extends whatever termui.Plot object this uses.
	// Add download speed plot
	progplot := widgets.NewPlot()
	progplot.Title = "Download Speed"
	progplot.Marker = widgets.MarkerBraille
	progplot.Data = [][]float64{
		[]float64{0, 0}, // Last observed download speed datum.
		[]float64{0, 0}, // Average download speed from last period.
	}
	progplot.AxesColor = ui.ColorWhite
	progplot.LineColors[0] = ui.ColorYellow
	progplot.DrawDirection = widgets.DrawRight
	statsRow := ui.NewRow(0.5, progplot)
	pd := &progressTracker{
		fromMoment:   time.Now(),
		fromProgress: int64(0),
	}
	updatePlot := func() { // TODO: refactor out updaters.
		read := tor.Stats().BytesReadUsefulData
    // Convert to updates per second.
		read64 := read.Int64() * updatesPerSecond
		// Convert to kB for graph.
		rotateIntoPlot(progplot, 0, math.RoundToEven(pd.getSpeedUpdate(read64)/1000))
		rotateIntoPlot(progplot, 1, func(ns []float64) float64 {
			s := float64(0)
			for _, n := range ns {
				s += n
			}
			return math.RoundToEven(float64(s) / float64(len(ns)))
		}(progplot.Data[0]))
		lastRate := int64(progplot.Data[1][len(progplot.Data[1])-1])
		progplot.Title = fmt.Sprintf(
			"Download Speed: %v • ETA: %v",
			// Convert to bytes per second.
			byteLengthToString(lastRate*1000)+"/s",
			func() string {
				if lastRate == 0 {
					return "∞"
				}
				toRead := tor.Length() - read64
				quantaRemaining := toRead / (lastRate * 1000)
				timeRemaining := time.Duration(quantum.Nanoseconds() * quantaRemaining)
				return timeRemaining.Round(time.Second).String()
			}(),
		)
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
