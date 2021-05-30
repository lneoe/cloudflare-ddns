package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	UPDATE_URL = "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s"
)

var config = struct {
	Token    string `json:"token"`
	ZoneID   string `json:"zone_id"`
	Domain   string `json:"domain"`
	RecordID string `json:"record_id"`
}{}

func loadConfig(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("read config file err: %v", err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("read file content error: %v", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("parse config err: %v", err)
	}
}

type LocalAddress struct {
	Ip string `json:"ip"`
}

type Result struct {
	Content string `json:"content"`
}

type Record struct {
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Messages []string `json:"messages"`
	Result   Result   `json:"result"`
}

type Form struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

func NewForm(ip string) *Form {
	return &Form{
		Type:    "A",
		Content: ip,
		ID:      config.RecordID,
		Name:    config.Domain,
	}
}

type Updater struct {
	url string

	client *http.Client
}

func NewUpdater() *Updater {
	url := fmt.Sprintf(UPDATE_URL, config.ZoneID, config.RecordID)
	cli := http.Client{
		Timeout: 10 * time.Second,
	}
	return &Updater{
		url:    url,
		client: &cli,
	}
}

func (updater Updater) setDefaultHeader(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Token))
}

func (updater Updater) getDNS() string {
	request, _ := http.NewRequest("GET", updater.url, nil)
	updater.setDefaultHeader(request)

	resp, err := updater.client.Do(request)
	if err != nil {
		log.Printf("get dns record failed, err: %v\n", err)
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("get dns record response Code: %d", resp.StatusCode)
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read response body error: %v", err)
		return ""
	}

	record := Record{}
	err = json.Unmarshal(body, &record)
	if err != nil {
		log.Printf("parse response error: %v", err)
		return ""
	}

	return record.Result.Content
}

func (updater Updater) setDNS(addr string) (ok bool) {
	form := NewForm(addr)
	formData, _ := json.Marshal(&form)

	body := bytes.NewReader(formData)
	request, _ := http.NewRequest("PUT", updater.url, body)
	updater.setDefaultHeader(request)
	resp, err := updater.client.Do(request)
	if err != nil {
		log.Printf("do post request error: %v", err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("parse response error: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Println("update request failed")
		log.Println(string(respBody))
		return
	}

	result := map[string]interface{}{}
	_ = json.Unmarshal(respBody, &result)

	success := result["success"]
	return success.(bool)
}

func (updater Updater) UpdateIfChanged(localRecord string) {
	lastRecord := updater.getDNS()
	log.Printf("cloudflare current record is: %q", lastRecord)

	if localRecord != "" && localRecord != lastRecord {
		done := updater.setDNS(localRecord)
		log.Printf("dns record updated: %v", done)
	}
}

func NewDetector(name string) Detective {
	switch name {
	case "ipify":
		return &IpifyAPIImpl{}
	case "ip-cmd":
		return &UseIpCmd{}
	}

	panic(fmt.Sprintf("unsupported detector %s", name))
}

func main() {
	filePath := flag.String("c", "ddns.sample.json", "json config file path")
	detectorName := flag.String("d", "ipify", "")
	intervalStr := flag.String("i", "5m", "interval, default: 5m")
	flag.Parse()

	interval, err := time.ParseDuration(*intervalStr)
	if err != nil {
		log.Panicf("interval format error %s", err)
	}

	if len(flag.Args()) > 0 && flag.Args()[0] == "version" {
		PrintVersion()
		os.Exit(0)
	}

	loadConfig(*filePath)

	ticker := time.NewTicker(interval)
	detector := NewDetector(*detectorName)
	updater := NewUpdater()

	for range ticker.C {
		loc, err := detector.Determine()
		if err != nil {
			log.Printf("error occurred: %s", err)
		} else {
			log.Printf("detemined public ip is: %q", loc)
			updater.UpdateIfChanged(loc)
		}
	}
}
