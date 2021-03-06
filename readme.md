# Build and run instructions
Assuming you have Go installed and have set your gopath. If you get a gopath not set error, see https://stackoverflow.com/questions/21001387/how-do-i-set-the-gopath-environment-variable-on-ubuntu-what-file-must-i-edit

```
git clone https://github.com/sea-erkin/goCertCheck.git
cd goCertCheck
go get golang.org/x/net/html
go build
./goCertCheck -u=urls -a
```

```
./goCertCheck -h
Usage of ./goCertCheck:
  -u string
    	-u list of urls in a new line delimited file. Url must be a valid URL
  -a bool
      -a (optional) active flag will try to connect to url on port 443 to see if active
  -o string
    	-o (optional) output format csv or json (default "csv")
  -t int
    	-t (optional) mintime as epoch will filter results to only show newer than mintime
 
 ```


# What does this thing do?

Checks cert.sh for certificates issued for a list of provided URLs as well as certificates issued for subdomains of the provided URLs. I.e. https://crt.sh/?q=%25.erkin.xyz

# How does it do this?

Provide a list of urls

For each url in list of urls, query cert.sh with a wildcard

Parse the cert.sh HTML output to extract the date and subdomain

If active flag set, this will try to connect to each of the newly found subdomains

Will save output as csv by default, can specify json if you wish

# Why would you want to use this?

To identify externally accessible assets for a given url or domain. Sometimes these assets may have been forgotten about or be brand new and not properly tested from a security standpoint, making them juicy targets for hackers. My goal is not to build a tool to make it easier for hackers to find juicy targets, but to build a tool that anyone can use to help get a handle on their external presence.

An easy way to find externally accessible assets is to simply identify subdomains for a domain. An easy way to identify subdomains is to look at certificates issued for that domain as you would want to use certificates if you want to use TLS to encrypt data in transit.

# Todo

1. Utilize goroutines and provide workers as a parameter to make requests faster. Have to test with cert.sh though as we would not want to overwhelm this kind organization which is providing us with data.
