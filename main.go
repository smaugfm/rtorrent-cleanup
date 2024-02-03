package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/smaugfm/go-rtorrent"
	"github.com/smaugfm/go-rtorrent/xmlrpc"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	cfg := parseConfig()

	torrents, err := cfg.rtorrent.GetTorrents(context.Background(), rtorrent.ViewMain)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting torrents: %v\n", err))
		return
	}
	slog.Info(fmt.Sprintf("There are %d torrents in the rTorrent instance", len(torrents)))
	for i, torrent := range torrents {
		slog.Info(fmt.Sprintf("Checking #%02d %s", i+1, torrent.Name))
		if !torrent.Completed {
			// still downloading
			continue
		}
		state, err := cfg.rtorrent.State(context.Background(), torrent)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error getting state for torrent %s. Skipping... %v\n", torrent.Name, err))
			continue
		}
		if state == 1 {
			// still seeding (i.e. ratio group criteria not met)
			continue
		}
		stateChanged, err := cfg.rtorrent.StateChanged(context.Background(), torrent)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error getting state_changed for torrent %s. Skipping...%v\n", torrent.Name, err))
		}
		if !time.Now().After(stateChanged.Add(*cfg.wait)) {
			// not seeding but not enough time has passed since it closed the torrent
			continue
		}
		status, err := cfg.rtorrent.GetStatus(context.Background(), torrent)
		var kbs = status.UpRate / 1024
		if int64(kbs) < *cfg.skipUlSpeedKbs {
			slog.Info(fmt.Sprintf("DELETE   #%02d %s", i+1, torrent.Name))
			if !cfg.dryRun {
				err := cfg.rtorrent.Delete(context.Background(), torrent)
				if err != nil {
					slog.Error("Error deleting torrent %s", torrent.Name)
				}
			}
		} else {
			slog.Info(fmt.Sprintf("Torrent %s is still seeding at %d KB/s > %d KB/s", torrent.Name, kbs, cfg.skipUlSpeedKbs))
		}
	}
	slog.Info("Done")
	slog.Info("\n")
}

type Config struct {
	rtorrent       *rtorrent.Client
	dryRun         bool
	wait           *time.Duration
	skipUlSpeedKbs *int64
}

func parseConfig() *Config {
	flag.Usage = usage
	username := flag.String("username", "", "HTTP Basic Authentication username")
	pass := flag.String("password", "", "HTTP Basic Authentication password")
	wait := flag.Duration("wait", time.Duration(1)*time.Hour, "Minimum duration time after finished torrent is deleted. Defaults to 1h")
	dryRun := flag.Bool("dry-run", false, "Do not actually delete torrents")
	skipUlSpeed := flag.Int64("skip-ul-speed", 1024, "Skips torrent deletion if at the time of check it has uploading speed more than in Kb/s. Defaults to 1024KB/s")

	flag.Parse()

	url := flag.Arg(0)
	if url == "" {
		flag.Usage()
		os.Exit(1)
		return nil
	}
	cfg := &rtorrent.Config{Addr: url}
	if username != nil {
		cfg.BasicUser = *username
	}
	if pass != nil {
		cfg.BasicPass = *pass
	}
	client := newRtorrentClient(cfg)

	return &Config{client, *dryRun, wait, skipUlSpeed}
}

func newRtorrentClient(cfg *rtorrent.Config) *rtorrent.Client {
	throttledTransport := NewThrottledTransport(1*time.Second, 1, http.DefaultTransport)
	httpClient := &http.Client{Transport: throttledTransport, Timeout: 5 * time.Second}
	xmlrpcClient := xmlrpc.NewClientWithHTTPClient(cfg.Addr, httpClient)
	xmlrpcClient.BasicUser = cfg.BasicUser
	xmlrpcClient.BasicPass = cfg.BasicPass
	return rtorrent.NewClientWithXmlrpcClient(cfg.Addr, xmlrpcClient)
}

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] url\n", os.Args[0])
	flag.PrintDefaults()
}
