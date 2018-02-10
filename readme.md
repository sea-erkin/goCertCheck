# Build and run instructions

```
git clone https://github.com/sea-erkin/goCertCheck.git
cd goCertCheck
go build
./goCertCheck -u=urls -a
```

# What does this thing do?

Checks cert.sh for certificates issued for a list of provided URLs as well as certificates issued for subdomains of the provided URLs

# How does it do this?

Provide a list of urls
For each url in list of urls, query cert.sh with a wildcard
Parse the cert.sh HTML output to extract the date and subdomain
If active flag set, this will try to connect to each of the newly found subdomains

# Todo

1. Utilize goroutines and provide workers as a parameter to make requests faster. Have to test with cert.sh though as we would not want to overwhelm this kind organization which is providing us with data.
