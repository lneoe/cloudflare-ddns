package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
)

type Detective interface {
	Determine() (string, error)
}

type UseIpCmd struct{}

func (u UseIpCmd) Determine() (string, error) {
	return u.getLocalAddress()
}

func (u UseIpCmd) getLocalAddress() (string, error) {
	cmd := exec.Command("/bin/sh", "-c", "ip addr show pppoe-wan | awk 'NR==3 {print $2}'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("get local address error: %v\n", err)
		return "", fmt.Errorf("get local address error: %w", err)
	}

	output = output[:len(output)-1]
	return string(output), nil
}

const ApiIpifyUrl = "https://api.ipify.org/?format=json"

// IpifyAPIImpl use https://api.ipify.org?format=json
type IpifyAPIImpl struct{}

func (impl IpifyAPIImpl) Determine() (string, error) {
	resp, err := http.Get(ApiIpifyUrl)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("check public address failed")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	address := LocalAddress{}
	err = json.Unmarshal(body, &address)
	if err != nil {
		log.Printf("unmarshal response body failed, err: %v\n", err)
	}

	return address.Ip, nil
}
