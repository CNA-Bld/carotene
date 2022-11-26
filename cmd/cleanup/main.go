package main

import (
	"context"
	"github.com/CNA-Bld/carotene/internal/utils"
	"github.com/otiai10/copy"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/sync/errgroup"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func has(m map[string]interface{}, k string) bool {
	_, ok := m[k]
	return ok
}

func main() {
	p, files := utils.ParsePathArg()

	for _, n := range []string{"TeamRaces", "RoomRaces", "CircleDetails"} {
		_ = os.MkdirAll(filepath.Join(p, n), os.ModeDir)
	}

	baseMoveToDir := filepath.Join(p, "Backup")

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(16)

	for _, f := range files {
		f := f
		g.Go(func() error {
			fileName := f.Name()
			if f.IsDir() || !strings.HasSuffix(fileName, ".msgpack") {
				return nil
			}

			mtSegs := []string{baseMoveToDir}
			if t, err := utils.TimeFromFileName(fileName); err == nil {
				mtSegs = append(mtSegs, t.Format("20060102"))
			}
			moveToPath := path.Join(mtSegs...)

			filePath := filepath.Join(p, fileName)
			if strings.HasSuffix(fileName, "R.msgpack") {
				info, err := f.Info()
				if err != nil {
					panic(err)
				}
				if info.Size() < 512*1024 {
					copyToSeg := ""

					b, err := os.ReadFile(filePath)
					if err != nil {
						panic(err)
					}
					d := make(map[string]interface{})
					if err := msgpack.Unmarshal(b, &d); err == nil {
						// Just continue if we fail to unmarshal. Some response packets adhere to msgpack specs but not JSON.
						if d, ok := d["data"]; ok {
							data := d.(map[string]interface{})

							if has(data, "race_start_params_array") && has(data, "race_result_array") && has(data, "rp_info") {
								// Team Races
								copyToSeg = "TeamRaces"
							} else if has(data, "room_info") && has(data, "race_horse_data_array") {
								// Champions Meeting
								copyToSeg = "RoomRaces"
							} else if has(data, "circle_info") && has(data, "circle_user_array") {
								// Circle Detail
								copyToSeg = "CircleDetails"
							}
						}
					}

					if copyToSeg != "" {
						err := copy.Copy(filePath, filepath.Join(p, copyToSeg, fileName), copy.Options{PreserveTimes: true})
						if err != nil {
							panic(err)
						}
					}
				}
			}

			_ = os.MkdirAll(moveToPath, os.ModeDir)
			err := os.Rename(filePath, filepath.Join(moveToPath, fileName))
			if err != nil {
				panic(err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
