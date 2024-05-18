package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	tld "github.com/jpillora/go-tld"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var (
	domain  = os.Getenv("DOMAIN")
	cfToken = os.Getenv("CLOUDFLARE_API_TOKEN")
)

func main() {
	api, err := cloudflare.NewWithAPIToken(cfToken)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to create cloudflare api client"))
	}

	type record struct {
		zoneID   string
		cfRecord cloudflare.DNSRecord
	}
	domains := lo.Reduce(strings.Split(domain, ","), func(recs []record, domain string, i int) []record {
		domain = strings.TrimSpace(domain)

		// The TLD package requires a domain with a scheme, otherwise output will be empty and no error will be returned
		u, err := tld.Parse("https://" + domain)
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to parse url"))
		}

		mainDomain := u.Domain + "." + u.TLD

		zoneID, err := api.ZoneIDByName(mainDomain)
		if err != nil {
			log.Fatal(errors.Wrapf(err, "failed to get zone id for domain %s", domain))
		}

		cfRecs, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{Name: domain})
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to list dns records"))
		}
		if len(cfRecs) == 0 {
			log.Fatal("no dns record found, please create it first")
		}
		theRecord, ok := lo.Find(cfRecs, func(r cloudflare.DNSRecord) bool {
			// Supports only A records for now
			return r.Name == domain && r.Type == "A"
		})
		if !ok {
			log.Fatalf("no A record found, please create it first, domain: %s", domain)
		}

		recs = append(recs, record{
			zoneID:   zoneID,
			cfRecord: theRecord,
		})
		return recs
	}, []record{})

	// Most API calls require a Context
	ctx := context.Background()

	log.Println("start watching ip changes")
	for {
		for _, domain := range domains {
			func() {
				log.Println("checking ip for domain", domain.cfRecord.Name)
				ip, err := getMyIP()
				if err != nil {
					log.Println(err)
					return
				}
				if ip == domain.cfRecord.Content {
					log.Println("ip not changed", domain.cfRecord.Name)
					return
				}
				log.Printf("ip for %s changed to %s\n", domain.cfRecord.Name, ip)

				domain.cfRecord, err = api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(domain.zoneID), cloudflare.UpdateDNSRecordParams{
					ID:      domain.cfRecord.ID,
					Type:    domain.cfRecord.Type,
					Name:    domain.cfRecord.Name,
					Content: ip,
					Proxied: domain.cfRecord.Proxied,
					TTL:     domain.cfRecord.TTL,
				})
				if err != nil {
					log.Println(errors.Wrapf(err, "failed to update dns record for domain %s", domain.cfRecord.Name))
					return
				}
				log.Println("dns record updated for domain", domain.cfRecord.Name)
			}()
		}
		time.Sleep(1 * time.Minute)
	}
}

func getMyIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", errors.Wrap(err, "failed to call ipify")
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read ipify response")
	}
	return string(ip), nil
}
