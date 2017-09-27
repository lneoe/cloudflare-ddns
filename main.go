package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var config = struct {
	ApiKey   string `json:"api_key"`
	Email    string `json:"email"`
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

type Address struct {
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

func publicAddress() string {
	url := "https://api.ipify.org/?format=json"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("check public address failed")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	address := Address{}
	err = json.Unmarshal(body, &address)
	if err != nil {
		log.Printf("unmarshal response body failed, err: %v\n", err)
	}

	return address.Ip
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
	url         string
	lastAddress string
}

func NewUpdater() *Updater {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s",
		config.ZoneID, config.RecordID)
	return &Updater{
		url: url,
	}
}

func (updater Updater) setDefaultHeader(r *http.Request) {
	r.Header.Set("X-Auth-Key", config.ApiKey)
	r.Header.Set("X-Auth-Email", config.Email)
}

func (updater Updater) getDNS() string {
	request, _ := http.NewRequest("GET", updater.url, nil)
	updater.setDefaultHeader(request)

	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("get dns record failed, err: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("get dns record response Code: %d", resp.StatusCode)
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read response body error: %v", err)
	}

	log.Println(string(body))

	record := Record{}
	err = json.Unmarshal(body, &record)
	if err != nil {
		log.Printf("parse response error: %v", err)
		return ""
	}

	return record.Result.Content
}

func (updater Updater) updateDNS() {

}

func main() {
	//publicAddress()

	loadConfig("ddns.dev.json")

}
