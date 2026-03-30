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
		go func() {
			notfound, _records, err := LookupMX(domain)
			records = append(records, ResponseDNSRecords[MXRecord]{
				Domain:   domain,
				Notfound: notfound,
				Records:  _records,
				Error:    err,
			})
			wg.Done()
		}()
	}

	wg.Wait()

	return records
}

func AsyncDNSLookupDKIM(selector string, domains []string) []ResponseDNSRecords[string] {
	if len(domains) == 0 {
		return nil
	}

	if selector == "" {
		selector = defaultSelector
	}

	var records []ResponseDNSRecords[string]
	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func() {
			notfound, _records, err := LookupDKIM(domain, selector)
			records = append(records, ResponseDNSRecords[string]{
				Domain:   fmt.Sprintf("%s._domainkey.%s", selector, domain),
				Notfound: notfound,
				Records:  _records,
				Error:    err,
			})
			wg.Done()
		}()
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
		go func() {
			notfound, _records, err := LookupDMARC(domain)
			records = append(records, ResponseDNSRecords[string]{
				Domain:   fmt.Sprintf("_dmarc.%s", domain),
				Notfound: notfound,
				Records:  _records,
				Error:    err,
			})
			wg.Done()
		}()
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
		go func() {
			notfound, _records, err := LookupSRV(domain, dnsType)
			records = append(records, ResponseDNSRecords[SRVRecord]{
				Domain:   fmt.Sprintf("_%s._tcp.%s", dnsType, domain),
				Notfound: notfound,
				Records:  _records,
				Error:    err,
			})
			wg.Done()
		}()
	}

	wg.Wait()

	return records
}

func AsyncDNSLookupRecursiveSPF(domains []string) []ResponseDNSRecords[string] {
	if len(domains) == 0 {
		return nil
	}

	var records []ResponseDNSRecords[string]

	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go func() {
			spf, totalQueries, err := LookupRecursiveSPF(domain, 0)
			notfound, e := IsDNSErrorNoSuchHost(err)
			if err == nil && len(spf) == 0 {
				notfound = true
			}
			records = append(records, ResponseDNSRecords[string]{
				Domain:       domain,
				Records:      spf,
				Notfound:     notfound,
				TotalQueries: totalQueries,
				Error:        e,
			})
			wg.Done()
		}()
	}

	wg.Wait()

	return records
}
