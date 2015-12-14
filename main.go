package main

//go:generate make bindata

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/Luzifer/rconfig"
	"github.com/fatih/color"
)

var (
	cfg = struct {
		FullScan         bool     `flag:"full-scan,f" default:"false" description:"Scan all nameservers included in this build"`
		Assert           []string `flag:"assert,a" default:"" description:"Exit with exit code 2 when these DNS entries were not found"`
		AssertPercentage float64  `flag:"assert-threshold" default:"100.0" description:"If used with -a fail when not at least N percent of the nameservers had the expected result"`
		Quiet            bool     `flag:"quiet,q" default:"false" description:"Do not communicate by text, use only exit codes"`
		Short            bool     `flag:"short,s" default:"true" description:"Use short notation (only when using assertion)"`
		Version          bool     `flag:"version" default:"false" description:"Print version and exit"`
	}{}
	nameserverDirectory = struct {
		CoreProviders     []string            `yaml:"core_providers"`
		PublicNameservers map[string][]string `yaml:"public_nameservers"`
	}{}
	version = "dev"

	// Color output helpers
	red         = color.New(color.FgRed).SprintfFunc()
	green       = color.New(color.FgGreen).SprintFunc()
	yellow      = color.New(color.FgYellow).SprintFunc()
	providerOut = color.New(color.FgWhite).Add(color.BgBlue).SprintfFunc()
	serverOut   = color.New(color.FgWhite).SprintfFunc()
)

type checkResult struct {
	Provider        string
	Server          string
	Results         []string
	QueryError      error
	AssertSucceeded bool
}

func (c checkResult) Print() {
	if c.QueryError != nil {
		fmt.Printf("%s %s %s\n",
			providerOut("[%s]", c.Provider),
			serverOut("(%s)", c.Server),
			red("Error: %s", c.QueryError),
		)
		return
	}

	var result string
	if len(cfg.Assert) > 0 {
		if c.AssertSucceeded {
			result = green("\u2713")
		} else {
			result = red("\u2717")
		}
	} else {
		result = ""
	}

	srvBuf := bytes.NewBuffer([]byte{})
	fmt.Fprintf(srvBuf, "%s %s %s\n",
		providerOut("[%s]", c.Provider),
		serverOut("(%s)", c.Server),
		result,
	)
	if !cfg.Short {
		for _, r := range c.Results {
			fmt.Fprintf(srvBuf, " %s %s\n", yellow("-"), r)
		}
	}
	fmt.Print(srvBuf.String())
}

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse arguments: %s", err)
	}

	if cfg.Version {
		fmt.Printf("dns_check version %s\n", version)
		os.Exit(0)
	}

	if reflect.DeepEqual(cfg.Assert, []string{""}) {
		cfg.Short = false
		cfg.Assert = []string{}
	}

	if err := loadNameservers(); err != nil {
		log.Fatalf("Unable to load nameserver list, probably your build is defect: %s", err)
	}
}

func main() {
	args := rconfig.Args()
	if len(args) != 3 {
		fmt.Println("Usage: dns_check <type> <query>")
		os.Exit(1)
	}

	queryType := args[1]
	queryFQDN := args[2]

	// Correct ordering is required for DeepEqual
	sort.Strings(cfg.Assert)

	wg := make(chan bool, 10)
	results := []*checkResult{}
	for provider, servers := range nameserverDirectory.PublicNameservers {
		if !cfg.FullScan && !isCoreProvider(provider) {
			continue
		}

		for _, server := range servers {
			wg <- true
			r := &checkResult{
				Provider: provider,
				Server:   server,
			}
			results = append(results, r)
			go checkProviderServer(wg, queryType, queryFQDN, provider, server, r)
		}
	}
	for len(wg) > 0 {
		time.Sleep(1)
	}

	var failCount int
	for _, r := range results {
		if !r.AssertSucceeded {
			failCount++
		}
		if !cfg.Quiet {
			r.Print()
		}
	}

	if (1.0-float64(failCount)/float64(len(results)))*100 < cfg.AssertPercentage {
		os.Exit(2)
	}
}

func checkProviderServer(wg chan bool, queryType, queryFQDN, provider, server string, r *checkResult) {
	r.Results, r.QueryError = getDNSQueryResponse(queryType, queryFQDN, server)

	if len(cfg.Assert) > 0 {
		r.AssertSucceeded = reflect.DeepEqual(r.Results, cfg.Assert)
	} else {
		r.AssertSucceeded = true
	}

	<-wg
}

func isCoreProvider(s string) bool {
	for _, v := range nameserverDirectory.CoreProviders {
		if v == s {
			return true
		}
	}
	return false
}

func loadNameservers() error {
	data, err := Asset("nameservers.yaml")
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &nameserverDirectory)
}
