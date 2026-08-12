package dnsutils

import (
	"fmt"
	"sync"
)

func AsyncDNSLookupMX(domains []string) []ResponseDNSRecords[MXRecord] {
	if len(domains) == 0 {
		return nil
	}

	var records []ResponseDNSRecords[MXRecord]
	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := LookupMX(d)
			records = append(records, ResponseDNSRecords[MXRecord]{
				Domain:   d,
				Notfound: notfound,
				Records:  _records,
				Error:    err,
			})
		}(domain)
	}

	wg.Wait()

	return records
}

func AsyncDNSLookupDKIM(selector string, domains []string) []ResponseDNSRecords[string] {
	if len(domains) == 0 {
		return nil
	}

	var records []ResponseDNSRecords[string]
	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := LookupDKIM(d, selector)
			records = append(records, ResponseDNSRecords[string]{
				Domain:   fmt.Sprintf("%s._domainkey.%s", selector, d),
				Notfound: notfound,
				Records:  _records,
				Error:    err,
			})
		}(domain)
	}

	wg.Wait()

	return records
}

func AsyncDNSLookupDMARC(domains []string) []ResponseDNSRecords[string] {
	if len(domains) == 0 {
		return nil
	}

	var records []ResponseDNSRecords[string]
	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := LookupDMARC(d)
			records = append(records, ResponseDNSRecords[string]{
				Domain:   fmt.Sprintf("_dmarc.%s", d),
				Notfound: notfound,
				Records:  _records,
				Error:    err,
			})
		}(domain)
	}

	wg.Wait()

	return records
}

func AsyncDNSLookupSRV(domains []string, dnsType string) []ResponseDNSRecords[SRVRecord] {
	if len(domains) == 0 {
		return nil
	}

	var records []ResponseDNSRecords[SRVRecord]
	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			notfound, _records, err := LookupSRV(d, dnsType)
			records = append(records, ResponseDNSRecords[SRVRecord]{
				Domain:   fmt.Sprintf("_%s._tcp.%s", dnsType, d),
				Notfound: notfound,
				Records:  _records,
				Error:    err,
			})
		}(domain)
	}

	wg.Wait()

	return records
}

func AsyncDNSLookupRecursiveSPF(domains []string) (records []ResponseDNSRecords[string]) {
	if len(domains) == 0 {
		return
	}

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
			records = append(records, ResponseDNSRecords[string]{
				Domain:       d,
				Records:      spf,
				Notfound:     notfound,
				TotalQueries: totalQueries,
				Error:        e,
			})
		}(domain)
	}

	wg.Wait()

	return
}
