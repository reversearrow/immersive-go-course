package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	mcrouterFlag   = flag.String("mcrouter", "", "enter the port mcrouter is running on")
	memcachedsFlag = flag.String("memcacheds", "", "enter comma seperated ports for the memcached servers")

	errMCRouterEmpty = errors.New("empty value entered for the mcrouterFlag")
)

// TODO: add additional validation
func flagsParseAndValidate() error {
	flag.Parse()
	if *mcrouterFlag == "" {
		return errMCRouterEmpty
	}
	return nil
}

func main() {
	logger := log.New(os.Stdout, "memcached-clusters: ", log.LstdFlags)

	if err := flagsParseAndValidate(); err != nil {
		logger.Printf("flags validation error: %v", err)
		os.Exit(1)
	}

	mcRouterClient := memcache.New(*mcrouterFlag)
	if err := mcRouterClient.Ping(); err != nil {
		logger.Printf("error pining the mcrouterFlag: %v", err)
		os.Exit(1)
	}

	const (
		key        = "number"
		value      = "8"
		expiration = 240
	)

	if err := mcRouterClient.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: expiration,
	}); err != nil {
		logger.Printf("error setting mcRouter item: %v", err)
		os.Exit(1)
	}

	topology := "replicated"
	var found bool
	memcachedServers := strings.Split(*memcachedsFlag, ",")
	for _, v := range memcachedServers {
		c := memcache.New(v)
		item, err := c.Get(key)
		if err != nil {
			switch {
			case errors.Is(err, memcache.ErrCacheMiss):
				topology = "sharded"
				continue
			default:
				logger.Printf("unexpected error occurred: %v", err)
				os.Exit(1)
			}
		}

		if item != nil {
			found = true
		}
	}

	if !found {
		logger.Printf("internal server occurred, item not found")
		os.Exit(1)
	}

	if topology == "replicated" {
		logger.Printf("memcache topology is replicated")
		os.Exit(0)
	}

	logger.Printf("memcache topology is sharded")
}
