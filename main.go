package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/rudi9719/loggy"
)

type ipRange struct {
	start net.IP
	end   net.IP
}

var privateRanges = []ipRange{
	ipRange{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	ipRange{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	ipRange{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	ipRange{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("182.168.255.255"),
	},
	ipRange{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}

var target string
var log loggy.Logger
var proofText string

func init() {
	// create the options struct and set globals
	var opts loggy.LogOpts
	opts.ProgName = "proof-bot"
	opts.Level = 5
	opts.UseStdout = true

	// if the keybase team was provided, enable it
	keybase_team := os.Getenv("PROOF_LOG_TEAM")
	if keybase_team != "" {
		opts.KBTeam = keybase_team
		keybase_channel := os.Getenv("PROOF_LOG_CHANNEL")
		if keybase_channel != "" {
			opts.KBChann = keybase_channel
		}
	}

	// if the log file path was provided enable it
	proof_log_file := os.Getenv("PROOF_LOG_FILE")
	if proof_log_file != "" {
		opts.OutFile = proof_log_file
	}

	// create the logger
	log = loggy.NewLogger(opts)

	// get the proof (keybase.txt)
	b, err := ioutil.ReadFile("/home/keybase/proof/keybase.txt")
	if err != nil {
		log.LogPanic("Unable to read proof file from /home/keybase/proof/keybase.txt\nExiting...")
	}
	proofText = string(b)

	// ensure that the redirect url was provided
	target = os.Getenv("PROOF_TARGET_URL")
	if target == "" {
		log.LogPanic("Unable to find PROOF_TARGET_URL in environment. Exiting.")
	}
}

func main() {
	http.HandleFunc("/", ServeRequest)
	http.HandleFunc("/keybase.txt", ServeProof)
	http.HandleFunc("/.well-known/keybase.txt", ServeProof)
	http.ListenAndServe(":8080", nil)
}

func ServeRequest(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, target, 302)
	clientIP := getClientIPAddress(r)
	clientUA := r.UserAgent()
	log.LogInfo(fmt.Sprintf("%s %s", clientIP, clientUA))
}

func ServeProof(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", proofText)
	clientIP := getClientIPAddress(r)
	clientUA := r.UserAgent()
	log.LogInfo(fmt.Sprintf("%s %s", clientIP, clientUA))
}

func inRange(r ipRange, ipAddress net.IP) bool {
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
		return true
	}
	return false
}

func isPrivateSubnet(ipAddress net.IP) bool {
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		for _, r := range privateRanges {
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}

func getClientIPAddress(r *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-IP"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		// scan from right to left until we find a public address
		// that will be the unproxied address
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
				// bad address
				continue
			}
			return ip
		}
	}
	// unable to find a real public ip address, so just return whats in the request
	// r.RemoteAddr returns "1.2.3.4:1234" format, so split on the colon becuase we don't care about originating port
	rip := strings.Split(r.RemoteAddr, ":")
	return rip[0]
}
