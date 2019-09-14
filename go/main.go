package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"sync"
)

const Delta = 0.5
const Eps = Delta / 10.

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var tunnelsFilename string
	flag.StringVar(&tunnelsFilename, "tunnel-path", "", "Path to the discretized tunnel (Required)")
	flag.Parse()

	if tunnelsFilename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	tunnel := LoadTunnelFromPdbFile(tunnelsFilename)
	disks := generateInitialDisks(tunnel)
	optimizeDisks(tunnel, disks)

	dumpResult(disks, tunnelsFilename+".disks")
}

func generateInitialDisks(tunnel Tunnel) []Disk {
	centers := []Vec3{}
	for _, sphere := range tunnel.Spheres {
		centers = append(centers, sphere.center)
	}
	dirs := []Vec3{}
	for i := 0; i < len(centers)-1; i++ {
		c1, c2 := centers[i], centers[i+1]
		dirs = append(dirs, SubVec3(c2, c1))
	}

	getNormal := func(cIdx int, offset float64) Vec3 {
		d := dirs[cIdx].Length()
		if offset < d/2 {
			n1, n2 := dirs[MaxInt(cIdx-1, 0)].Normalized(), dirs[cIdx].Normalized()
			w := 0.5 + offset/d
			return AddVec3(n1.Scaled(1-w), n2.Scaled(w)).Normalized()
		}
		n1, n2 := dirs[cIdx].Normalized(), dirs[MinInt(cIdx+1, len(dirs)-1)].Normalized()
		w := 1 - (offset/d - 0.5)
		return AddVec3(n1.Scaled(w), n2.Scaled(1-w)).Normalized()
	}

	type Job struct {
		idx                int
		center, baseNormal Vec3
		result             *Disk
	}

	tasksCh := make(chan Job)
	finishedJobsCh := make(chan Job, 10000)

	wg := new(sync.WaitGroup)
	for i := 1; i < 8; i++ {
		wg.Add(1)
		go func() {
			for job := range tasksCh {
				minDisk := tunnel.getMinimalDisk(job.center, job.baseNormal)
				job.result = &minDisk
				finishedJobsCh <- job
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(finishedJobsCh)
	}()

	disks := []Disk{}
	l := 0.
	for cIdx := 0; cIdx < len(centers)-1; cIdx++ {
		fmt.Println(cIdx)
		c1, c2 := centers[cIdx], centers[cIdx+1]
		v := SubVec3(c2, c1)
		dst := v.Length()
		nDir := v.Normalized()

		for l < dst {
			center := AddVec3(c1, nDir.Scaled(l))
			normal := getNormal(cIdx, l)
			tasksCh <- Job{
				idx:        len(disks),
				center:     center,
				baseNormal: normal,
			}
			disks = append(disks, Disk{})
			l += Delta
		}
		l = l - dst
	}
	close(tasksCh)
	for job := range finishedJobsCh {
		disks[job.idx] = *job.result
	}
	return disks
}

func dumpResult(disks []Disk, filePath string) {
	file, err := os.Create(filePath)
	panicOnError(err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	floatToString := func(input_num float64) string {
		return strconv.FormatFloat(input_num, 'f', 6, 64)
	}

	for _, disk := range disks {
		err := writer.Write([]string{
			floatToString(disk.center.x),
			floatToString(disk.center.y),
			floatToString(disk.center.z),
			floatToString(disk.normal.x),
			floatToString(disk.normal.y),
			floatToString(disk.normal.z),
			floatToString(disk.radius),
		})
		panicOnError(err)
	}
}
