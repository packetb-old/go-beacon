package main

import (
	"bitbucket.org/gdamore/mangos"
	"bitbucket.org/gdamore/mangos/protocol/rep"
	"bitbucket.org/gdamore/mangos/transport/all"
	"flag"
	"fmt"
	"github.com/ugorji/go/codec"
	"reflect"
	"strconv"
	"time"
)

var (
	mh codec.MsgpackHandle
	b  []byte
)

func boomerangMetrics(d map[string][]string) {
	fmt.Println("------ server msg ------ ")
	nt_dns, _ := delta(d["nt_dns_st"][0], d["nt_dns_end"][0])                               // domainLookupEnd - domainLookupStart
	nt_con, _ := delta(d["nt_con_st"][0], d["nt_con_end"][0])                               // connectEnd - connectStart
	nt_domcontloaded, _ := delta(d["nt_domcontloaded_st"][0], d["nt_domcontloaded_end"][0]) // domContentLoadedEnd - domContentLoadedStart
	nt_processed, _ := delta(d["nt_domcontloaded_st"][0], d["nt_domcomp"][0])               // domComplete - domContentLoadedStart
	nt_request, _ := delta(d["nt_req_st"][0], d["nt_res_st"][0])                            // ResponseStart - RequestStart
	nt_response, _ := delta(d["nt_res_st"][0], d["nt_res_end"][0])                          // ResponseEnd - ResponseStart
	nt_navtype := d["nt_nav_type"][0]
	roundtrip, _ := delta(d["rt.bstart"][0], d["rt.end"][0])
	page := d["r"][0]
	url := d["u"][0]

	fmt.Println("Navigation type: ", nt_navtype)
	fmt.Println("Navigation timing DNS: ", nt_dns)
	fmt.Println("Navigation timing Connection: ", nt_con)
	fmt.Println("Navigation timing DOM content loaded: ", nt_domcontloaded)
	fmt.Println("Navigation timing DOM processing: ", nt_processed)
	fmt.Println("Navigation timing Request: ", nt_request)
	fmt.Println("Navigation timing Response: ", nt_response)
	fmt.Println("Roundtrip: ", roundtrip)
	fmt.Println("Page: ", page)
	fmt.Println("URL: ", url)
	fmt.Println("------ server msg ------ ")
}

func jsMetrics(d map[string][]string) {
	fmt.Println("------ server msg ------ ")
	nt_dns, _ := delta(d["nt_dns_st"][0], d["nt_dns_end"][0])                               // domainLookupEnd - domainLookupStart
	nt_con, _ := delta(d["nt_con_st"][0], d["nt_con_end"][0])                               // connectEnd - connectStart
	nt_domcontloaded, _ := delta(d["nt_domcontloaded_st"][0], d["nt_domcontloaded_end"][0]) // domContentLoadedEnd - domContentLoadedStart
	nt_processed, _ := delta(d["nt_domcontloaded_st"][0], d["nt_domcomp"][0])               // domComplete - domContentLoadedStart
	nt_request, _ := delta(d["nt_req_st"][0], d["nt_res_st"][0])                            // ResponseStart - RequestStart
	nt_response, _ := delta(d["nt_res_st"][0], d["nt_res_end"][0])                          // ResponseEnd - ResponseStart
	nt_navtype := d["nt_nav_type"][0]
	roundtrip, _ := delta(d["rt.bstart"][0], d["rt.end"][0])
	page := d["r"][0]
	url := d["u"][0]

	fmt.Println("Navigation type: ", nt_navtype)
	fmt.Println("Navigation timing DNS: ", nt_dns)
	fmt.Println("Navigation timing Connection: ", nt_con)
	fmt.Println("Navigation timing DOM content loaded: ", nt_domcontloaded)
	fmt.Println("Navigation timing DOM processing: ", nt_processed)
	fmt.Println("Navigation timing Request: ", nt_request)
	fmt.Println("Navigation timing Response: ", nt_response)
	fmt.Println("Roundtrip: ", roundtrip)
	fmt.Println("Page: ", page)
	fmt.Println("URL: ", url)
	fmt.Println("------ server msg ------ ")

}

// Calculate delta between start and end
func delta(start string, end string) (int, error) {
	s, err := strconv.Atoi(start)
	if err != nil {
		return -1, err
	}
	e, err := strconv.Atoi(end)
	if err != nil {
		return -1, err
	}
	return e - s, nil
}

func decode(buf []byte) (error, map[string][]string) {

	doc := map[string][]string(nil)
	dec := codec.NewDecoderBytes(buf, &mh)
	err := dec.Decode(&doc)

	if err != nil {
		return err, nil
	}
	return nil, doc
}

func main() {
	// consumer -type boomerang -listen tcp://127.0.0.1:8000 -statsd 192.168.33.20:8125
	var (
		listenAddr   string
		statsdServer string
		trackerType  string
	)

	flag.StringVar(&listenAddr, "listen", "tcp://127.0.0.1:8000", "Listening string - default: tcp://127.0.0.1:8000")
	flag.StringVar(&statsdServer, "statsd", "127.0.0.1:8125", "statsd endpoint - default: 127.0.0.1:8125")
	flag.StringVar(&trackerType, "tracker", "boomerang", "tracker type - default: boomerang [boomerang, js]")

	responseServerReady := make(chan struct{})
	responseServer, err := rep.NewSocket()
	defer responseServer.Close()

	all.AddTransports(responseServer)
	if err != nil {
		fmt.Println("Error connecting: ", err)
		return
	}
	fmt.Println("Listening:", listenAddr)
	fmt.Println("Statsd endpoint:", statsdServer)
	fmt.Println("Tracker type: ", trackerType)
	fmt.Println("Consumer ready")

	go func() {
		var err error
		var serverMsg *mangos.Message

		if err = responseServer.Listen(listenAddr); err != nil {
			fmt.Printf("\nServer listen failed: %v", err)
			return
		}

		close(responseServerReady)
		mh.MapType = reflect.TypeOf(map[string][]string(nil))

		for {
			if serverMsg, err = responseServer.RecvMsg(); err != nil {
				fmt.Printf("\nServer receive failed: %v", err)
			}
			err, d := decode(serverMsg.Body)
			if len(d) < 1 {
				fmt.Println("Discarded message")
				continue
			}
			switch trackerType {
			case "boomerang":
				boomerangMetrics(d)
			case "js":
				jsMetrics(d)
			}

			serverMsg.Body = []byte("OK")
			if err = responseServer.SendMsg(serverMsg); err != nil {
				fmt.Printf("\nServer send failed: %v", err)
				return
			}
		}
		fmt.Println("Listening")
	}()

	for {
		time.Sleep(10 * time.Second)
	}
}
