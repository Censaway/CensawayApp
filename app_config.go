package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strconv"
)

func (a *App) generateConfig(vlessLink string) (string, error) {
	u, err := url.Parse(vlessLink)
	if err != nil {
		return "", fmt.Errorf("bad link")
	}

	q := u.Query()
	uuid := u.User.Username()
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())

	transportType := q.Get("type")
	if transportType == "" {
		transportType = "tcp"
	}
	security := q.Get("security")

	sni := q.Get("sni")
	pbk := q.Get("pbk")
	sid := q.Get("sid")
	fp := q.Get("fp")
	if fp == "" {
		fp = "chrome"
	}

	path := q.Get("path")
	hostHeader := q.Get("host")
	serviceName := q.Get("serviceName")
	flow := q.Get("flow")

	vlessOutbound := map[string]interface{}{
		"type":            "vless",
		"tag":             "proxy",
		"server":          host,
		"server_port":     port,
		"uuid":            uuid,
		"flow":            flow,
		"packet_encoding": "xudp",
	}

	tlsConfig := map[string]interface{}{
		"enabled":     false,
		"server_name": sni,
		"utls": map[string]interface{}{
			"enabled":     true,
			"fingerprint": fp,
		},
	}

	if security == "tls" {
		tlsConfig["enabled"] = true
		vlessOutbound["tls"] = tlsConfig
	} else if security == "reality" {
		tlsConfig["enabled"] = true
		tlsConfig["reality"] = map[string]interface{}{
			"enabled":    true,
			"public_key": pbk,
			"short_id":   sid,
		}
		vlessOutbound["tls"] = tlsConfig
	}

	transportConfig := map[string]interface{}{
		"type": transportType,
	}

	if transportType == "ws" {
		transportConfig["path"] = path
		if hostHeader != "" {
			transportConfig["headers"] = map[string]string{"Host": hostHeader}
		}
	} else if transportType == "grpc" {
		transportConfig["service_name"] = serviceName
		if serviceName == "" {
			transportConfig["service_name"] = path
		}
	} else if transportType == "http" {
		transportConfig["host"] = []string{hostHeader}
		transportConfig["path"] = path
	}

	if transportType != "tcp" {
		vlessOutbound["transport"] = transportConfig
	}

	outbounds := []map[string]interface{}{
		vlessOutbound,
		{"type": "direct", "tag": "direct"},
	}

	inbounds := []map[string]interface{}{}

	inbounds = append(inbounds, map[string]interface{}{
		"type":        "mixed",
		"tag":         "mixed-in",
		"listen":      "127.0.0.1",
		"listen_port": a.Settings.MixedPort,
		"sniff":       true,
	})

	if a.Settings.RunMode == "tun" {
		tunConfig := map[string]interface{}{
			"type":                       "tun",
			"tag":                        "tun-in",
			"address":                    []string{"172.19.0.1/30"},
			"mtu":                        9000,
			"auto_route":                 true,
			"strict_route":               true,
			"stack":                      "system",
			"sniff":                      true,
			"sniff_override_destination": true,
		}

		if runtime.GOOS == "linux" {
			tunConfig["interface_name"] = "tun0"
		}

		inbounds = append(inbounds, tunConfig)
	}

	ruleSets := []map[string]interface{}{}
	if a.Settings.RoutingMode == "smart" {
		ruleSets = append(ruleSets, map[string]interface{}{
			"tag":             "geoip-ru",
			"type":            "remote",
			"format":          "binary",
			"url":             "https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set/geoip-ru.srs",
			"download_detour": "proxy",
		})
	}

	rules := []map[string]interface{}{}

	rules = append(rules, map[string]interface{}{
		"protocol": "dns",
		"action":   "hijack-dns",
	})

	if a.Settings.RunMode == "tun" {
		rules = append(rules, map[string]interface{}{
			"inbound": "tun-in",
			"action":  "sniff",
		})
	}

	for _, ur := range a.Settings.UserRules {
		r := map[string]interface{}{}
		if ur.Outbound == "block" {
			r["action"] = "reject"
		} else {
			r["action"] = "route"
			r["outbound"] = ur.Outbound
		}

		if ur.Type == "domain" {
			r["domain_suffix"] = []string{ur.Value}
		} else if ur.Type == "ip" {
			r["ip_cidr"] = []string{ur.Value}
		} else if ur.Type == "process" {
			r["process_name"] = []string{ur.Value}
		}
		rules = append(rules, r)
	}

	rules = append(rules, map[string]interface{}{
		"ip_is_private": true,
		"action":        "route",
		"outbound":      "direct",
	})

	if a.Settings.RoutingMode == "smart" {
		if len(a.Settings.RuDomains) > 0 {
			rules = append(rules, map[string]interface{}{
				"domain_suffix": a.Settings.RuDomains,
				"action":        "route",
				"outbound":      "direct",
			})
		}

		rules = append(rules, map[string]interface{}{
			"rule_set": "geoip-ru",
			"action":   "route",
			"outbound": "direct",
		})
	}

	if ip := net.ParseIP(host); ip != nil {
		rules = append(rules, map[string]interface{}{
			"ip_cidr":  []string{host + "/32"},
			"action":   "route",
			"outbound": "direct",
		})
	} else {
		rules = append(rules, map[string]interface{}{
			"domain":   []string{host},
			"action":   "route",
			"outbound": "direct",
		})
	}

	rules = append(rules, map[string]interface{}{
		"ip_cidr":  []string{"8.8.8.8/32", "1.1.1.1/32"},
		"action":   "route",
		"outbound": "direct",
	})

	rules = append(rules, map[string]interface{}{
		"inbound":  "clash-api",
		"action":   "route",
		"outbound": "direct",
	})

	dnsRules := []map[string]interface{}{}
	
	if a.Settings.RoutingMode == "smart" {
		dnsRules = append(dnsRules, map[string]interface{}{
			"domain_suffix": a.Settings.RuDomains,
			"server":        "local_dns",
		})
	}

	dnsConfig := map[string]interface{}{
		"servers": []map[string]interface{}{
			{
				"tag":    "remote_dns",
				"type":   "udp",
				"server": "8.8.8.8",
				"detour": "proxy",
			},
			{
				"tag":  "local_dns",
				"type": "local",
			},
		},
		"rules":    dnsRules,
		"final":    "remote_dns",
		"strategy": "ipv4_only",
	}

	fullConfig := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"experimental": map[string]interface{}{
			"clash_api": map[string]interface{}{
				"external_controller": "127.0.0.1:9090",
			},
			"cache_file": map[string]interface{}{
				"enabled":    true,
				"store_rdrc": true,
			},
		},
		"dns":       dnsConfig,
		"inbounds":  inbounds,
		"outbounds": outbounds,
		"route": map[string]interface{}{
			"rule_set":                ruleSets,
			"rules":                   rules,
			"auto_detect_interface":   true,
			"final":                   "proxy",
			"default_domain_resolver": "local_dns",
		},
	}

	bytes, err := json.MarshalIndent(fullConfig, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
