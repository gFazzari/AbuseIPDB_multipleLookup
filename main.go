package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://api.abuseipdb.com/api/v2/check"

type AbuseIpDbResp struct {
	Data struct {
		IpAddress            string
		IsPublic             bool
		IpVersion            int
		IsWhitelisted        bool
		AbuseCondidenceScore int
		CountryCode          string
		UsageType            string
		Isp                  string
		Domain               string
		Hostnames            []string
		TotalReports         int
		NumDistinctUsers     int
		LastReportedAt       string
	}
}

func main() {

	var key string
	fmt.Println("Insert your API Key: ")
	if _, err := fmt.Scanln(&key); err != nil {
		log.Fatal("...Error while reading API Key...")
	}

	fmt.Println("Reading a list of IPs or domain in file ip.txt.")
	fmt.Println("Each element must be in a separate line.")
	f1, err := os.Open("./ip.txt")
	if err != nil {
		log.Fatalln("...Cant't open ip.txt...")
	}
	defer f1.Close()

	f2, err := os.Create("res.txt")
	if err != nil {
		log.Fatalln("...Cant't create new file to write results...")
	}

	scanner := bufio.NewScanner(f1)
	for scanner.Scan() {
		element := scanner.Text()
		temp_element, err := domainToIP(element)
		if err != nil {
			fmt.Printf("Element %s could not have been evaluated.\n", element)
			continue
		}
		temp := evaluate(key, temp_element)
		if temp.Data.AbuseCondidenceScore > 0 || temp.Data.TotalReports > 0 {
			_, err = f2.WriteString(fmt.Sprintf("%s --> %+v\n\n", element, temp.Data))
			if err != nil {
				log.Fatal("...Cant't write in output file...")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
}

func evaluate(key string, query string) AbuseIpDbResp {
	// Create an HTTP client to make the request
	client := &http.Client{Timeout: time.Duration(5) * time.Second}
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		log.Fatalln("...Error in creating the request...")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Key", key)
	q := req.URL.Query()
	q.Add("ipAddress", query)
	q.Add("maxAgeInDays", "90")
	req.URL.RawQuery = q.Encode()
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln("...Error during request...")
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("...Error during body lecture...")
	}
	defer res.Body.Close()

	var jsonResp AbuseIpDbResp
	if err = json.Unmarshal(b, &jsonResp); err != nil {
		log.Fatal("...Cannot unmarshal AbuseIpDb response...")
	}

	return jsonResp
}

func domainToIP(domain string) (string, error) {
	ips, err := net.LookupHost(domain)
	if err != nil {
		return "", err
	}

	var ret []string
	for _, ip := range ips {
		if ipv4 := net.ParseIP(ip); ipv4 != nil {
			ret = append(ret, ip)
		}
	}
	if len(ret) > 0 {
		return ret[0], nil
	} else {
		return "", errors.New("...Cannot validate IPv4...")
	}
}
