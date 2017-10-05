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
	"os/exec"
	"time"
)

const (
	UPDATE_URL         = "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s"
	DETECT_ADDRESS_URL = "https://api.ipify.org/?format=json"
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

//func (updater Updater) getLocalAddress() string {
//	resp, err := http.Get(DETECT_ADDRESS_URL)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if resp.StatusCode != http.StatusOK {
//		log.Println("check public address failed")
//	}
//
//	body, _ := ioutil.ReadAll(resp.Body)
//
//	address := LocalAddress{}
//	err = json.Unmarshal(body, &address)
//	if err != nil {
//		log.Printf("unmarshal response body failed, err: %v\n", err)
//	}
//
//	return address.Ip
//}

func (updater Updater) getLocalAddress() string {
	cmd := exec.Command("/bin/sh", "-c", "ip addr show pppoe-wan | awk 'NR==3 {print $2}'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("get local address error: %v\n", err)
		return ""
	}

	output = output[:len(output)-1]
	return string(output)
}

func (updater Updater) setDefaultHeader(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Auth-Key", config.ApiKey)
	r.Header.Set("X-Auth-Email", config.Email)
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
	json.Unmarshal(respBody, &result)

	success := result["success"]
	return success.(bool)
}

func (updater Updater) Run() {
	for {
		localAddress := updater.getLocalAddress()
		log.Printf("local pubilic address is: %s", localAddress)

		lastRecord := updater.getDNS()
		log.Printf("current record is: %s", lastRecord)

		if localAddress != "" && localAddress != lastRecord {
			done := updater.setDNS(localAddress)
			log.Printf("dns record updated: %v", done)
		}

		time.Sleep(1 * time.Minute)
	}
}

func main() {
	filePath := flag.String("c", "ddns.json", "ddns.json")
	flag.Parse()

	loadConfig(*filePath)

	updater := NewUpdater()
	updater.Run()
}
