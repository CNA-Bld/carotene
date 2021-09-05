package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"github.com/CNA-Bld/carotene/internal/utils"
	"github.com/vmihailenco/msgpack/v5"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

type row struct {
	name   string
	counts map[time.Time]uint64
}

func main() {
	d := flag.Duration("duration", 60*24*time.Hour, "Duration to generate report")
	tcp := flag.Int64("circle", 0, "Circle ID to record; 0 (default) for circles the packet capturer is in; -1 for all")
	circleID := uint64(*tcp)

	p, files := utils.ParsePathArg()
	since := time.Now().Add(-*d)

	rows := make(map[uint64]*row)
	var times []time.Time

	for _, f := range files {
		t, err := utils.TimeFromFileName(f.Name())
		if err != nil || since.After(t) {
			continue
		}

		b, err := os.ReadFile(filepath.Join(p, f.Name()))
		if err != nil {
			panic(err)
		}

		d := make(map[string]interface{})
		decoder := msgpack.NewDecoder(bytes.NewReader(b))
		decoder.UseLooseInterfaceDecoding(true)
		if err := decoder.Decode(&d); err == nil {
			data := d["data"].(map[string]interface{})

			if circleID >= 0 {
				if circleID > 0 && circleID != data["circle_info"].(map[string]interface{})["circle_id"].(uint64) {
					continue
				}

				capturerViewerID := d["data_headers"].(map[string]interface{})["viewer_id"].(uint64)

				hasCapturer := false
				for _, ud := range data["summary_user_info_array"].([]interface{}) {
					user := ud.(map[string]interface{})
					if capturerViewerID == user["viewer_id"].(uint64) {
						hasCapturer = true
						break
					}
				}
				if !hasCapturer {
					continue
				}
			}

			times = append(times, t)

			for _, ud := range data["summary_user_info_array"].([]interface{}) {
				user := ud.(map[string]interface{})
				viewerID := user["viewer_id"].(uint64)
				if _, ok := rows[viewerID]; !ok {
					rows[viewerID] = &row{counts: make(map[time.Time]uint64)}
				}
				rows[viewerID].name = user["name"].(string)
				rows[viewerID].counts[t] = user["fan"].(uint64)
			}
		}
	}

	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	header := []string{"ID", "Name"}
	for _, t := range times {
		header = append(header, t.Format(time.RFC3339))
	}
	records := [][]string{header}

	for viewerId, row := range rows {
		r := []string{strconv.Itoa(int(viewerId)), row.name}
		for _, t := range times {
			if c, ok := row.counts[t]; ok {
				r = append(r, strconv.Itoa(int(c)))
			} else {
				r = append(r, "")
			}
		}
		records = append(records, r)
	}

	f, err := os.Create(time.Now().Format("20060102") + ".csv")
	defer f.Close()
	if err != nil {
		panic(err)
	}

	w := csv.NewWriter(f)
	if err := w.WriteAll(records); err != nil {
		panic(err)
	}
}
