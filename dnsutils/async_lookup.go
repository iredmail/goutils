package dnsutils

import (
	"fmt"
	"sync"
)

func AsyncDNSLookupMX(domains []string) []ResponseDNSRecords[MXRecord] {
	if len(domains) == 0 {
		return nil
	}

	chanRecords := make(chan ResponseDNSRecords[MXRecord], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := LookupMX(d)
			chanRecords <- ResponseDNSRecords[MXRecord]{
				Domain:   d,
				Notfound: notfound,
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

func AsyncDNSLookupDKIM(selector string, domains []string) []ResponseDNSRecords[string] {
	if len(domains) == 0 {
		return nil
	}

	chanRecords := make(chan ResponseDNSRecords[string], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := LookupDKIM(d, selector)
			chanRecords <- ResponseDNSRecords[string]{
				Domain:   fmt.Sprintf("%s._domainkey.%s", selector, d),
				Notfound: notfound,
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

func AsyncDNSLookupDMARC(domains []string) []ResponseDNSRecords[string] {
	if len(domains) == 0 {
		return nil
	}

	chanRecords := make(chan ResponseDNSRecords[string], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := LookupDMARC(d)
			chanRecords <- ResponseDNSRecords[string]{
				Domain:   fmt.Sprintf("_dmarc.%s", d),
				Notfound: notfound,
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

func AsyncDNSLookupSRV(domains []string, dnsType string) []ResponseDNSRecords[SRVRecord] {
	if len(domains) == 0 {
		return nil
	}

	chanRecords := make(chan ResponseDNSRecords[SRVRecord], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := LookupSRV(d, dnsType)
			chanRecords <- ResponseDNSRecords[SRVRecord]{
				Domain:   fmt.Sprintf("_%s._tcp.%s", dnsType, d),
				Notfound: notfound,
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

func AsyncDNSLookupRecursiveSPF(domains []string) (records []ResponseDNSRecords[string]) {
	if len(domains) == 0 {
		return
	}

	chanRecords := make(chan ResponseDNSRecords[string], len(domains))

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			spf, totalQueries, err := LookupRecursiveSPF(d, 0)
			notfound, e := IsDNSErrorNoSuchHost(err)
			if err == nil && len(spf) == 0 {
				notfound = true
			}

			chanRecords <- ResponseDNSRecords[string]{
				Domain:       d,
				Records:      spf,
				Notfound:     notfound,
				TotalQueries: totalQueries,
				Error:        e,
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
