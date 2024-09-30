cloudprefixes is a lightweight tool designed to assist in recon by handling IP prefixes published by cloud and hosting providers. The tool automatically retrieves these prefixes, stores them in an SQLite database, and offers a straightforward interface for querying the database, either through the command line or by integrating with other security tools.

Key Features:
- Service Association: Associate a managed service to an IP address assist in focus of further recon.
- Prefix Management: Fetch and update IP prefixes from various cloud platforms (e.g., AWS, GitHub) with a single command, ensuring your database stays up to date.
- Query Functionality: Search for cloud provider information associated with one or multiple IP addresses, either as command-line arguments or via standard input.
- Extensibility: The SQLite database allows easy integration with third-party tools and scripts, making it flexible for security automation and further recon use cases.
- Prettify Output: Easily pipe the output into tools like jq for clean, human-readable JSON formatting.


# Useage

```
$ cloudprefixes -h

    __ _      ___  __ __ ___   ____  ____    ___ _____ ____ __ __ 
   /  | T    /   \|  T  |   \ |    \|    \  /  _|     l    |  T  T
  /  /| |   Y     |  |  |    \|  o  |  D  )/  [_|   __j|  T|  |  |
 /  / | l___|  O  |  |  |  D  |   _/|    /Y    _|  l_  |  |l_   _j
/   \_|     |     |  :  |     |  |  |    \|   [_|   _] |  ||     |
\     |     l     l     |     |  |  |  .  |     |  T   j  l|  |  |
 \____l_____j\___/ \__,_l_____l__j  l__j\_l_____l__j  |____|__j__|

Usage
  cloudprefixes [OPTION]... [IP ADDRESS]...
Search cloud prefixes in database for each IP ADDRESS

With no IP ADDRESS, read standard input.

Options:
  -dbpath string
    	path to database file (default "./cloudprefixes.db")
  -update
    	update all prefixes in database and exit

```

Before being able to query the database, it needs to be populated with the ranges by executing the following
```
$ cloudprefixes -update
```

Querying can be multiple IP addresses as arguments or piped to stdin
```
$ ./cloudprefixes 192.30.252.1 2600:1f13:0a0d:a700::1
[{"prefix":"192.30.252.0/22","platform":"GitHub","service":"Hooks"},{"prefix":"192.30.252.0/22","platform":"GitHub","service":"Web"},{"prefix":"192.30.252.0/22","platform":"GitHub","service":"API"},{"prefix":"192.30.252.0/22","platform":"GitHub","service":"Git"},{"prefix":"192.30.252.0/22","platform":"GitHub","service":"GithubEnterpriseImporter"},{"prefix":"192.30.252.0/22","platform":"GitHub","service":"Copilot"}]
[{"prefix":"2600:1f13::/36","platform":"AWS","region":"us-west-2","service":"AMAZON","metadata":"{\"network_boarder_group\":\"us-west-2\"}"},{"prefix":"2600:1f13::/36","platform":"AWS","region":"us-west-2","service":"EC2","metadata":"{\"network_boarder_group\":\"us-west-2\"}"},{"prefix":"2600:1f13:a0d:a700::/56","platform":"AWS","region":"us-west-2","service":"EC2_INSTANCE_CONNECT","metadata":"{\"network_boarder_group\":\"us-west-2\"}"}]
```

Piping the output to jq will prettify it
```
$ ./cloudprefixes 2600:1f13:0a0d:a700::1 |jq
[
  {
    "prefix": "2600:1f13::/36",
    "platform": "AWS",
    "region": "us-west-2",
    "service": "AMAZON",
    "metadata": "{\"network_boarder_group\":\"us-west-2\"}"
  },
  {
    "prefix": "2600:1f13::/36",
    "platform": "AWS",
    "region": "us-west-2",
    "service": "EC2",
    "metadata": "{\"network_boarder_group\":\"us-west-2\"}"
  },
  {
    "prefix": "2600:1f13:a0d:a700::/56",
    "platform": "AWS",
    "region": "us-west-2",
    "service": "EC2_INSTANCE_CONNECT",
    "metadata": "{\"network_boarder_group\":\"us-west-2\"}"
  }
]
```

The database is SQLite so can be queried directly
```
$ sqlite3 cloudprefixes.db 
SQLite version 3.45.1 2024-01-30 16:01:20
Enter ".help" for usage hints.
sqlite> select service, count(prefix) from cloud_prefixes where platform is "GitHub" and ip_version = 6 group by service;
API|2
Actions|862
Copilot|2
Git|2
GithubEnterpriseImporter|2
Hooks|2
Pages|4
Web|2
```


# Prefixes
## Major Cloud

The major cloud providers publish the most useful information about their prefixes. They include what service the prefixes are associated with. Each provider has their own format requiring a custom parser for each one.

### Oracle
- https://docs.oracle.com/en-us/iaas/tools/public_ip_ranges.json

Tag details from https://docs.oracle.com/en-us/iaas/Content/General/Concepts/addressranges.htm

Valid tag values:
 - `OCI`: The VCN CIDR blocks
 - `OSN`: The CIDR block ranges for the Oracle Services Network.
 - `OBJECT_STORAGE`: The CIDR block ranges used by the Object Storage service. For more information, see Overview of Object Storage. 

### Google
- GCP - https://www.gstatic.com/ipranges/cloud.json
- Google APIs - https://www.gstatic.com/ipranges/goog.json

### AWS
- https://ip-ranges.amazonaws.com/ip-ranges.json

### Azure
- Public - https://www.microsoft.com/en-us/download/details.aspx?id=56519
- US Government - https://www.microsoft.com/en-us/download/details.aspx?id=57063
- China - https://www.microsoft.com/en-us/download/details.aspx?id=57062
- Germany - https://www.microsoft.com/en-au/download/details.aspx?id=57064

Service tag details: https://learn.microsoft.com/en-us/azure/virtual-network/service-tags-overview

### GitHub
- https://api.github.com/meta

https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-githubs-ip-addresses


## Geofeed

Geofeed is a list of self-published IP address ranges with geolocation information. The standard for geofeed is published under [RFC8805](https://datatracker.ietf.org/doc/html/rfc8805) and contains the following fields:
- IP Prefix (required)
- Alpha2code
- Region
- City
- Postcode

For example:
```
# Constant.com / Vultr.com GeoFeed (AS20473)
# Email: support@vultr.com
# Last Updated: 2024-09-29 13:38:43
8.3.29.0/24,US,US-CA,Los Angeles,90012
8.6.8.0/24,US,US-CA,Los Angeles,90012
8.6.193.0/24,US,US-FL,Miami,33142
8.9.3.0/24,US,US-NJ,Piscataway,08854
8.9.4.0/24,US,US-NJ,Piscataway,08854
8.9.5.0/24,US,US-NJ,Piscataway,08854
```

CloudFlare is an example of a provider who publishes a list with only prefixes. It is unclear if they intended to publish to RFC8805 but this can still be process with CSV parser used for geofeeds.

There many geofeeds published and more can easily be added. For now the sources are a few minor cloud providers

## Digital Ocean
- https://digitalocean.com/geo/google.csv

## Linode
- https://geoip.linode.com/

## Vultr
- https://geofeed.constant.com/

## CloudFlare
- https://www.cloudflare.com/ips-v4
- https://www.cloudflare.com/ips-v6

# License

This project is licensed under the GPLv3 License - see the LICENSE file for details