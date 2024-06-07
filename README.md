# RSEScan

RSEScan is a command-line utility for interacting with the [RSECloud API](https://rsecloud.com/ "RSECloud"). It allows you to fetch subdomains and IPs from certificates for a given domain or organization.

## Features

- Fetch subdomains for a given domain.
- Fetch IPs from certificates for a given domain.
- Fetch IPs from certificates for a given organization.

## Installation

To install RSEScan, run:

```sh
go install github.com/shdwpwn/rsescan@latest
```

## Usage
### Fetch Subdomains

To fetch subdomains for a given domain:

```sh
rsescan -d example.com -key YOUR_API_KEY
```
If the API key is not provided, the utility will attempt to read it from `~/.config/rsescan/api_key`.

### Fetch IPs from Certificates by Domain

To fetch certificates for a given domain:

```sh
rsescan -d example.com -cn -key YOUR_API_KEY
```
### Fetch IPs from Certificates by Organization

To fetch certificates for a given organization:

```sh
rsescan -so "Organization Name" -key YOUR_API_KEY
```

## Examples
### Fetch Subdomains Example

```sh
rsescan -d example.com
```

Output:
```makefile
www.example.com
vpn.example.com
test-dev.example.com
```
### Fetch IPs from Certificates by Domain Example

```sh
rsescan -d example.com -cn
```

Output:
```makefile
192.168.1.1:443
192.168.1.2:443
192.168.1.3:443
```
### Fetch IPs from Certificates by Organization Example

```sh
rsescan -so "Example Org"
```

Output:

```makefile
192.168.1.1:443
192.168.1.2:443
192.168.1.3:443
```
## Configuration

If you do not wish to pass the API key with every command, you can save it to a file. Create a file at `~/.config/rsescan/api_key` and put your API key in it.

```sh
mkdir -p ~/.config/rsescan
echo "YOUR_API_KEY" > ~/.config/rsescan/api_key
```
## License
This project is licensed under the MIT License.
