package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ammario/ipisp/v2"
	"github.com/zenthangplus/goccm"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func checkProxy(addr string, t int) {
	defer c.Done()

	timeout := time.Duration(t) * time.Second
	var proxyDialer, err = proxy.SOCKS5("tcp", addr, &proxy.Auth{}, &net.Dialer{
		Timeout: timeout,
	})

	if err != nil {
		return
	}

	client := &http.Client{
		Timeout: 2 * timeout,
		Transport: &http.Transport{
			Dial: proxyDialer.Dial,
		},
	}

	resp, err := client.Get("http://ifconfig.me/")
	if err != nil {
		return
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	bodyString := string(bodyBytes)
	addrSlice := strings.Split(addr, ":")

	if bodyString == addrSlice[0] {

		data := map[string]interface{}{
			"ip":    addrSlice[0],
			"port":  addrSlice[1],
			"type":  "SOCKS",
			"proto": "tcp",
		}

		if resp, err := ipisp.LookupIP(context.Background(), net.ParseIP(bodyString)); err == nil {
			data["ISP"] = resp.ISPName
			data["country"] = resp.Country
			data["ASN"] = resp.ASN
			data["registry"] = resp.Registry
			data["address_allocation_time"] = resp.AllocatedAt
		}

		if domain, err := net.LookupAddr(bodyString); err == nil {
			data["domains"] = domain
		}

		p, _ := json.Marshal(data)
		fmt.Printf("%s\n", p)

	}
}

var c goccm.ConcurrencyManager

func main() {
	port := flag.String("p", "1080", "port")
	timeout := flag.Int("t", 5, "timeout")
	goroutines := flag.Int("r", 1000, "goroutines")
	flag.Parse()

	c = goccm.New(*goroutines)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		address := scanner.Text()

		c.Wait()
		if strings.Contains(address, ":") {
			go checkProxy(address, *timeout)
		} else {
			go checkProxy(scanner.Text()+":"+*port, *timeout)

		}
	}
	c.WaitAllDone()
}
