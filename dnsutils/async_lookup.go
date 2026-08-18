package dnsutils

import (
	"fmt"
	"sync"
)

func AsyncDNSLookupMX(r Resolver, domains []string) []ResponseDNSRecords[MXRecord] {
	if len(domains) == 0 {
		return nil
	}

	chanRecords := make(chan ResponseDNSRecords[MXRecord], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := r.LookupMX(d)
			chanRecords <- ResponseDNSRecords[MXRecord]{
				Domain:   d,
				Notfound: notfound || len(_records) == 0,
				Records:  _records,
				Error:    err,
			}
		}(domain)
	}

	wg.Wait()
	close(chanRecords)

	var records []ResponseDNSRecords[MXRecord]
	for record := range chanRecords {
		records = append(records, record)
	}

	return records
}

func AsyncDNSLookupDKIM(r Resolver, selector string, domains []string) []ResponseDNSRecords[string] {
	if len(domains) == 0 {
		return nil
	}

	chanRecords := make(chan ResponseDNSRecords[string], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := r.LookupDKIM(d, selector)
			chanRecords <- ResponseDNSRecords[string]{
				Domain:   fmt.Sprintf("%s._domainkey.%s", selector, d),
				Notfound: notfound || len(_records) == 0,
				Records:  _records,
				Error:    err,
			}
		}(domain)
	}

	wg.Wait()
	close(chanRecords)

	var records []ResponseDNSRecords[string]
	for record := range chanRecords {
		records = append(records, record)
	}

	return records
}

func AsyncDNSLookupDMARC(r Resolver, domains []string) []ResponseDNSRecords[string] {
	if len(domains) == 0 {
		return nil
	}

	chanRecords := make(chan ResponseDNSRecords[string], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := r.LookupDMARC(d)
			chanRecords <- ResponseDNSRecords[string]{
				Domain:   fmt.Sprintf("_dmarc.%s", d),
				Notfound: notfound || len(_records) == 0,
				Records:  _records,
				Error:    err,
			}
		}(domain)
	}

	wg.Wait()
	close(chanRecords)

	var records []ResponseDNSRecords[string]
	for record := range chanRecords {
		records = append(records, record)
	}

	return records
}

func AsyncDNSLookupSRV(r Resolver, domains []string, dnsType string) []ResponseDNSRecords[SRVRecord] {
	if len(domains) == 0 {
		return nil
	}

	chanRecords := make(chan ResponseDNSRecords[SRVRecord], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := r.LookupSRV(d, dnsType)
			chanRecords <- ResponseDNSRecords[SRVRecord]{
				Domain:   fmt.Sprintf("_%s._tcp.%s", dnsType, d),
				Notfound: notfound || len(_records) == 0,
				Records:  _records,
				Error:    err,
			}
		}(domain)
	}

	wg.Wait()
	close(chanRecords)

	var records []ResponseDNSRecords[SRVRecord]
	for record := range chanRecords {
		records = append(records, record)
	}

	return records
}

func AsyncDNSLookupRecursiveSPF(r Resolver, domains []string) (records []ResponseDNSRecords[string]) {
	if len(domains) == 0 {
		return
	}

	chanRecords := make(chan ResponseDNSRecords[string], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, spf, totalQueries, err := r.LookupRecursiveSPF(d, 0)
			chanRecords <- ResponseDNSRecords[string]{
				Domain:       d,
				Records:      spf,
				Notfound:     notfound || len(spf) == 0,
				TotalQueries: totalQueries,
				Error:        err,
			}
		}(domain)
	}

	wg.Wait()
	close(chanRecords)

	for record := range chanRecords {
		records = append(records, record)
	}

	return
}
