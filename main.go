package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/kr/pretty"
	"github.com/markcaudill/go-proxmox/proxmox"
)

func main() {
	username := flag.String("username", "root@pam", "the username to authenticate with")
	password := flag.String("password", "root", "the password to authenticate with")
	path := flag.String("path", "", "the path to query (e.g. /version)")
	flag.Parse()

	var credentials = map[string]string{
		"username": *username,
		"password": *password,
	}
	apiURL, err := url.Parse("https://127.0.0.1:8006/api2/json")
	if err != nil {
		log.Fatal("unable to parse URL: ", err)
	}
	client := resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	s, err := proxmox.NewSession(client, apiURL, credentials)
	if err != nil {
		log.Fatal(err)
	}

	res, err := s.Do("GET", *path, proxmox.QueryParams{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("GET %s: %# v", *path, pretty.Formatter(res))
}
