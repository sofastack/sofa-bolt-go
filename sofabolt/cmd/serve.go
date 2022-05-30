// nolint
// Copyright 20xx The Alipay Authors.
//
// @authors[0]: bingwu.ybw(bingwu.ybw@antfin.com|detailyang@gmail.com)
// @authors[1]: robotx(robotx@antfin.com)
//
// *Legal Disclaimer*
// Within this source code, the comments in Chinese shall be the original, governing version. Any comment in other languages are for reference only. In the event of any conflict between the Chinese language version comments and other language version comments, the Chinese language version shall prevail.
// *法律免责声明*
// 关于代码注释部分，中文注释为官方版本，其它语言注释仅做参考。中文注释可能与其它语言注释存在不一致，当中文注释与其它语言注释存在不一致时，请以中文注释为准。
//
//

package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fatih/color"
	"github.com/paulbellamy/ratecounter"
	"github.com/spf13/cobra"
	"github.com/sofastack/sofa-bolt-go/sofabolt"
)

var (
	echo       bool
	cpuprofile string
	memprofile string
	delay      int
	flush      int
	counter    *ratecounter.RateCounter
)

func init() {
	fs := serverCmd.Flags()
	fs.BoolVar(&echo, "echo", false, "Set the echo mode")
	fs.IntVar(&delay, "delay", 0, "Set the delay time in milliseconds")
	fs.IntVar(&flush, "flush", 0, "Set the max flush delay time in milliseconds")
	fs.StringVarP(&cpuprofile, "cpuprofile", "", "", "Set the cpuprofile file to write")
	fs.StringVarP(&memprofile, "memprofile", "", "", "Set the memprofile file to write")
}

func echohandler(rw sofabolt.ResponseWriter, req *sofabolt.Request) {
	counter.Incr(1)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	rw.GetResponse().SetContent(req.GetContent())
	if _, err := rw.Write(); err != nil {
		log.Println("failed to write", err)
	}
}

func handler(rw sofabolt.ResponseWriter, req *sofabolt.Request) {
	fmt.Println(req.String())
	res := rw.GetResponse()
	res.SetContent(req.GetContent())
	_, err := rw.Write()
	if err != nil {
		log.Println("failed to write", err)
	}
}

var serverCmd = &cobra.Command{
	Use:   "serve <address>",
	Short: "serve a bolt server and echo back to",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		counter = ratecounter.NewRateCounter(15 * time.Second)
		go func() {
			timer := time.NewTicker(15 * time.Second)
			defer timer.Stop()
			for range timer.C {
				fmt.Printf("Latest QPS:%d\n", counter.Rate()/15)
			}
		}()
		go func() {
			color.Black("try to run pprof server in :6060")
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()

		if cpuprofile != "" {
			var f *os.File
			f, err := os.Create(cpuprofile)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			// nolint
			defer f.Close()
			if err = pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
			defer pprof.StopCPUProfile()
			color.Black("> Write cpuprofile to %s", cpuprofile)
		}

		if memprofile != "" {
			var f *os.File
			f, err := os.Create(memprofile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			// nolint
			defer f.Close()
			runtime.GC() // get up-to-date statistics
			if err = pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			color.Black("> Write memprofile to %s", memprofile)
		}

		var (
			srv *sofabolt.Server
			err error
		)
		if echo {
			srv, err = sofabolt.NewServer(sofabolt.WithServerHandler(
				sofabolt.HandlerFunc(echohandler)),
				sofabolt.WithServerMaxPendingCommands(10240),
				sofabolt.WithServerMaxConnctions(10240),
				sofabolt.WithServerTimeout(0, 0, 0, time.Duration(flush)*time.Millisecond),
			)
		} else {
			srv, err = sofabolt.NewServer(sofabolt.WithServerHandler(
				sofabolt.HandlerFunc(handler)))
		}
		if err != nil {
			log.Fatal(err)
		}

		ln, err := net.Listen("tcp", args[0])
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Listening on", ln.Addr().String())
		if err = srv.Serve(ln); err != nil {
			log.Fatal(err)
		}
	},
}
