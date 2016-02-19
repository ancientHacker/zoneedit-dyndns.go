// zoneedit-dyndns.go - a dynamic dns updater for ZoneEdit
// Copyright (C) 2016 Daniel C. Brotsky.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
// Licensed under the LGPL v3.  See the LICENSE file for details

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

const (
	defaultEndpoint = "https://dynamic.zoneedit.com/dyn/jsclient.php"
)

type queryParam struct {
	name         string
	value        string
	defaultValue string
	flagName     string
	envName      string
	description  string
}

var (
	params = []*queryParam{
		&queryParam{name: "rsp_ident", defaultValue: "zoneedit",
			flagName: "service", envName: "DYNDNS_SERVICE",
			description: "DNS service hosting the domain",
		},
		&queryParam{name: "hostname", defaultValue: "myhost.example.com",
			flagName: "hostname", envName: "DYNDNS_HOSTNAME",
			description: "fully qualified hostname to update",
		},
		&queryParam{name: "wildcard", defaultValue: "NO",
			flagName: "wildcard", envName: "DYNDNS_WILDCARD",
			description: "specify YES to update the domain wildcard host",
		},
	}
	endpoint     = flag.String("endpoint", defaultEndpoint, "API `endpoint` to contact")
	username     = flag.String("uname", "", "authenticate as `username`")
	password     = flag.String("pword", "", "authenticate with `password`")
	showResponse = flag.Bool("s", true, "show server response")
)

func init() {
	for _, p := range params {
		flag.StringVar(&p.value, p.flagName, p.defaultValue, p.description)
	}
}

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		usage()
		os.Exit(2)
	}
	if *username == "" || *password == "" {
		fmt.Fprintf(os.Stderr, "error: You must specify a username and password.\n")
		usage()
		os.Exit(3)
	}
	query := ""
	for i, p := range params {
		if i == 0 {
			query += "?"
		} else {
			query += "&"
		}
		if p.name == "hostname" && p.value == p.defaultValue {
			fmt.Fprintf(os.Stderr, "error: You must specify a hostname.\n")
			usage()
			os.Exit(4)
		}
		if p.name == "wildcard" && p.value != "YES" && p.value != "NO" {
			fmt.Fprintf(os.Stderr, "error: wildcard must be YES or NO.\n")
			usage()
			os.Exit(5)
		}
		query += p.name + "=" + url.QueryEscape(p.value)
	}
	req, err := http.NewRequest("GET", *endpoint+query, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: request construction failed: %v\n", err)
		os.Exit(6)
	}
	req.SetBasicAuth(*username, *password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: request failed: %v\n", err)
		os.Exit(7)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: couldn't read response body: %v\n", err)
		os.Exit(8)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "error: response status was: %v\n", resp.Status)
		fmt.Fprintf(os.Stderr, "\tresponse body was: %v\n", string(body))
		os.Exit(9)
	}
	if *showResponse {
		fmt.Println(string(body))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: zoneedit-dyndns [flags] where flags are:\n")
	flag.PrintDefaults()
}
