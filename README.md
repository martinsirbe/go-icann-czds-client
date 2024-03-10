# ICANN CZDS client
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmartinsirbe%2Fgo-icann-czds-client.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmartinsirbe%2Fgo-icann-czds-client?ref=badge_shield)


A client library for the Internet Corporation for Assigned Names and Numbers (ICANN) Centralized Zone Data Service (CZDS).
The client aims to streamline the process of querying the CZDS to retrieve zone files and to list the Top-Level Domains (TLDs) 
available on your account. It supports in-memory JWT storage and features a mechanism to refetch the token if it doesn't exist, 
is found to be invalid, or has expired.  

Zone files obtained through this client contain a registry's list of registered and actively managed domain names, alongside 
various DNS records tailored to the registry's offered services. These files adhere to a specific format, initially defined 
in [RFC 1035](https://rfc-annotations.research.icann.org/rfc1035.html) and further refined by subsequent RFCs. However, 
the CZDS distributes these files in a format that conforms to a subset of these standards, as specified by 
the Zone File Access Advisory Group in their Strategy Proposal (Section 5.1.7, Page 9) and included in Specification 4 
of the Registry Agreement. This ensures users receive comprehensively structured data, enabling effective analysis and 
application of the information contained within TLD zone files.  

## Features

- **JWT Authentication**: Manages JWT tokens with an in-memory store, automatically refetching tokens as needed.
  It is also possible to supply a custom JWT token store by implementing the `TokenStore` interface. The custom
  token store can be provided via `TokenStoreOpt` when initialising a new client.
- **TLD Listing**: Enables the listing of TLDs available to your account, including the approval status for each TLD.
- **Zone File Queries**: Allows for the retrieval of zone files from CZDS, presenting the data organized by domain name
  along with associated DNS records.

## Prerequisites

Before using the client, you must obtain credentials from ICANN's Centralized Zone Data Service (CZDS). These credentials 
are necessary to authenticate and interact with the CZDS API.  

Follow the steps below to obtain your ICANN CZDS credentials:
1. Register for an ICANN Account:
    - Visit the [ICANN CZDS website](https://czds.icann.org/).
    - Click on the "Sign Up" or "Register" option to create a new ICANN account.
    - Follow the registration process, providing the required personal and contact information.
    - Submit your registration. You may need to verify your email address as part of this process.
2. Apply for TLD zone file access:
    - Once your ICANN account is active, log in to the [CZDS portal](https://czds.icann.org/home).
    - Navigate to the "Zone File Access" section.
    - Submit an application for zone file access. This will typically involve selecting the Top-Level Domains (TLDs) for 
      which you're requesting access and agreeing to the terms and conditions.
    - Your application will be reviewed by the respective TLD operators. Approval times may vary depending on the operator.
3. Accessing the CZDS API:
    - To access the API, you will need to use your username (your email) and password when initialising a new client.

Remember, your access to zone files is governed by the terms and conditions you agreed to during the application process. 
Ensure your use of the data complies with these terms.

## Usage

Import `go-icann-czds-client` into your Go project:

```go
import "github.com/yourusername/go-icann-czds-client/czds"
```

### Creating a Client

Initialise a new ICANN CZDS client with your credentials. By default, the client uses in-memory JWT storage:
```go
client := czds.NewClient("email", "your_password")
```

If you prefer to use a custom token store, implement the `TokenStore` interface and pass it when creating the client:
```go
client := icann.NewClient("email", "your_password", TokenStoreOpt(customTokenStore))
```

### Querying Zone File Data

To obtain zone file data for a specific TLD:
```go
zoneFile, err := client.GetZoneFileData("com")
if err != nil {
    log.Fatalf("failed to fetch zone file data: %v", err)
}
fmt.Println(zoneFile)
```

### Listing TLDs

To list TLDs:
```go
tlds, err := client.ListTLDs()
if err != nil {
    log.Fatalf("failed to list TLDs: %v", err)
}
fmt.Println(tlds)
```

## Contributing

Feel free to contribute to the project by submitting pull requests or creating issues for bugs and feature requests.

## License

This project is licensed under the MIT License. See [LICENSE.md](LICENSE.md).


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmartinsirbe%2Fgo-icann-czds-client.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmartinsirbe%2Fgo-icann-czds-client?ref=badge_large)

## Disclaimer

This client library is not officially affiliated with ICANN or the CZDS. Its purpose is to streamline access to data provided by the CZDS.