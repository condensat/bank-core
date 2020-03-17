// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"net"
	"net/http"

	"github.com/thoas/stats"
)

type StatsArgs struct {
}

type StatsService int

type StatsReply struct {
	Application string      `json:"application"`
	Version     string      `json:"version"`
	Host        string      `json:"host"`
	Statistics  *stats.Data `json:"statistics"`
}

var (
	StatsMiddleware = stats.New()
)

func (t *StatsService) Status(r *http.Request, args *StatsArgs, result *StatsReply) error {
	//	log.Printf("Statistics Check Called\n")
	reply := StatsReply{}
	reply.Application = "Authorization"
	reply.Host = getLocalIP()
	reply.Version = Version
	reply.Statistics = StatsMiddleware.Data()

	*result = reply
	return nil
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
