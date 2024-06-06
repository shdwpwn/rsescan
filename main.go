package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type SubdomainResponse struct {
	Count      int      `json:"count"`
	Data       []string `json:"data"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pagesize"`
	TotalPages int      `json:"total_pages"`
}

type CertificateResponse struct {
	Count      int               `json:"count"`
	Data       []CertificateData `json:"data"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pagesize"`
	TotalPages int               `json:"total_pages"`
}

type CertificateData struct {
	IP         string          `json:"ip"`
	Port       json.RawMessage `json:"port"`
	SubjectCN  string          `json:"subject_cn"`
	SubjectOrg string          `json:"subject_org"`
	Timestamp  string          `json:"timestamp"`
}

func fetchSubdomains(apiURL, apiKey, domain string, page int) (*SubdomainResponse, error) {
	url := fmt.Sprintf("%s?page=%d", apiURL, page)
	payload := strings.NewReader(fmt.Sprintf(`{"domain":"%s"}`, domain))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response SubdomainResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func fetchCertificates(apiURL, apiKey, query string, page int, queryType string) (*CertificateResponse, error) {
	url := fmt.Sprintf("%s?page=%d", apiURL, page)
	var payload *strings.Reader
	if queryType == "domain" {
		payload = strings.NewReader(fmt.Sprintf(`{"domain":"%s"}`, query))
	} else if queryType == "org" {
		payload = strings.NewReader(fmt.Sprintf(`{"org_name":"%s"}`, query))
	}

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response CertificateResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func parsePort(raw json.RawMessage) (int, error) {
	var portString string
	if err := json.Unmarshal(raw, &portString); err == nil {
		return strconv.Atoi(portString)
	}

	var portInt int
	if err := json.Unmarshal(raw, &portInt); err == nil {
		return portInt, nil
	}

	return 0, fmt.Errorf("unable to parse port: %s", raw)
}

func getAPIKeyFromFile() (string, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "rsescan", "api_key")
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	apiKey := strings.TrimSpace(string(data))
	if apiKey == "" {
		return "", fmt.Errorf("API key file is empty")
	}
	return apiKey, nil
}

func main() {
	domain := flag.String("d", "", "Domain to fetch subdomains or certificates for")
	apiKey := flag.String("key", "", "API key for authentication")
	useCN := flag.Bool("cn", false, "Use certificate search API")
	orgName := flag.String("so", "", "Organization name to search certificates by")

	flag.Parse()

	// Ensure only one of -d, -cn, or -so is used at a time and that -cn is used with -d
	if (*domain != "" && *orgName != "") || (*useCN && *orgName != "") || (*useCN && *domain == "") {
		log.Fatal("Invalid combination of flags. Use either -d, -cn, or -so, and -cn must be used with -d.")
	}

	if *domain == "" && *orgName == "" {
		log.Fatal("Domain or organization name must be provided.")
	}

	if *apiKey == "" {
		var err error
		*apiKey, err = getAPIKeyFromFile()
		if err != nil {
			log.Fatalf("API key not found: %v", err)
		}
	}

	var apiURL string
	var query string
	var queryType string

	if *useCN {
		apiURL = "https://api.rsecloud.com/api/v1/searchCertificatesByDomain"
		query = *domain
		queryType = "domain"
	} else if *orgName != "" {
		apiURL = "https://api.rsecloud.com/api/v1/searchCertificatesByOrgName"
		query = *orgName
		queryType = "org"
	} else {
		apiURL = "https://api.rsecloud.com/api/v1/subdomains"
		query = *domain
		queryType = "domain"
	}

	if queryType != "domain" || *useCN {
		allCertificates := []CertificateData{}

		initialResponse, err := fetchCertificates(apiURL, *apiKey, query, 1, queryType)
		if err != nil {
			log.Fatalf("Error fetching certificates: %v", err)
		}
		allCertificates = append(allCertificates, initialResponse.Data...)

		for page := 2; page <= initialResponse.TotalPages; page++ {
			response, err := fetchCertificates(apiURL, *apiKey, query, page, queryType)
			if err != nil {
				log.Fatalf("Error fetching certificates for page %d: %v", page, err)
			}
			allCertificates = append(allCertificates, response.Data...)
		}

		for _, certificate := range allCertificates {
			port, err := parsePort(certificate.Port)
			if err != nil {
				log.Printf("Error parsing port for certificate %s: %v", certificate.IP, err)
				continue
			}
			fmt.Printf("%s:%d\n", certificate.IP, port)
		}
	} else {
		allSubdomains := []string{}

		initialResponse, err := fetchSubdomains(apiURL, *apiKey, query, 1)
		if err != nil {
			log.Fatalf("Error fetching subdomains: %v", err)
		}
		allSubdomains = append(allSubdomains, initialResponse.Data...)

		for page := 2; page <= initialResponse.TotalPages; page++ {
			response, err := fetchSubdomains(apiURL, *apiKey, query, page)
			if err != nil {
				log.Fatalf("Error fetching subdomains for page %d: %v", page, err)
			}
			allSubdomains = append(allSubdomains, response.Data...)
		}

		for _, subdomain := range allSubdomains {
			fmt.Println(subdomain)
		}
	}
}
