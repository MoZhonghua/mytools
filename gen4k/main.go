package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

func WaitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}

func random_file_name(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		a := rand.Uint32()
		s += fmt.Sprintf("%02x", a%256)
	}

	return s
}

func gen_level_dirs_recursive(parent string, remain int) {
	if remain <= 0 {
		return
	}
	for i := 0; i < 256; i++ {
		s := fmt.Sprintf("%02x", i)
		d := filepath.Join(parent, s)
		os.MkdirAll(d, 0744)
		gen_level_dirs_recursive(d, remain-1)
	}
}

func gen_level_dirs(parent string) {
	gen_level_dirs_recursive(parent, level)
}

func random_size(min, max int) int {
	s := int(rand.Uint32())%(max-min) + min
	return s - (s % 1024)
}

func level_name(s string, level int) string {
	p := ""
	for i := 0; i < level; i++ {
		p = filepath.Join(p, s[2*i:2*(i+1)])
	}
	return filepath.Join(p, s)
}

func write_rand_file(parent string, size int) {
	s := random_file_name(16)
	p := filepath.Join(parent, level_name(s, level))

	f, err := os.Create(p)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	defer f.Close()
	f.Write(data[0:size])
}

var create_dir bool
var count int64
var parentDir string
var level int
var minSize int
var maxSize int

var data = make([]byte, 1024*1024*16)

func main() {
	flag.IntVar(&level, "level", 1, "levels of directory")
	flag.BoolVar(&create_dir, "mkdir", false, "create dir")
	flag.Int64Var(&count, "count", 0, "files to create")
	flag.IntVar(&maxSize, "max", 64, "max file size (KB)")
	flag.IntVar(&minSize, "min", 32, "min file size (KB)")
	flag.StringVar(&parentDir, "dir", ".", "parent dir")
	flag.Parse()

	if create_dir {
		gen_level_dirs(parentDir)
		return
	}

	rand.Read(data)
	created := int64(0)
	for i := 0; i < 5; i++ {
		go func() {
			for {
				if atomic.LoadInt64(&created) > count {
					return
				}
				write_rand_file(parentDir, random_size(minSize*1024, maxSize*1024))
				atomic.AddInt64(&created, 1)
			}
		}()
	}

	start := time.Now()
	last := start
	lastCreated := int64(0)
	lock := sync.Mutex{}
	go func() {
		for {
			time.Sleep(time.Second)
			lock.Lock()
			n := atomic.LoadInt64(&created)
			now := time.Now()
			avgSpeed := float64(n) / now.Sub(start).Seconds()
			lastSpeed := float64(n-lastCreated) / now.Sub(last).Seconds()
			last = now
			lastCreated = n
			log.Printf("%7d files @ %4.0f - %4.0f files/sec - %.0f seconds\n",
				n, avgSpeed, lastSpeed, now.Sub(start).Seconds())
			lock.Unlock()
			if n > count {
				break
			}
		}
	}()

	WaitForCtrlC()

	lock.Lock()
	time.Sleep(time.Second)
	n := atomic.LoadInt64(&created)
	now := time.Now()
	avgSpeed := float64(n) / now.Sub(start).Seconds()
	last = now
	lastCreated = n
	fmt.Println("==============================================")
	fmt.Printf("Total: %d files @ %.2f files/seconds - %.0f seconds\n",
		n, avgSpeed, now.Sub(start).Seconds())
	lock.Unlock()
}
