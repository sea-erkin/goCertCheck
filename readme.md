# Run instructions

  git clone
  cd dir
  go build
  ./certShQuery -h 

# What does this thing do?

Checks cert.sh for certificate issued for a wildcard subdomain for a list of provided urls.

Has the option to try to connect and allows filtering of a date range

# How does it do this?

Provide a list of urls

For each url in list of urls, query cert.sh
Parse the cert.sh output to extract the date and subdomain
If active flag set, this will try to connect

# Todo

1. Utilize goroutines and provide workers as a parameter to make requests faster. Have to test with cert.sh though as we would not want to overwhelm this kind organization which is providing us with data.