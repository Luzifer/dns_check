# Luzifer / dns\_check

dns\_check is a small utility to check major DNS services for records of a FQDN without having to query them one-by-one.

Use cases:

- Check whether the IP of your domain is consistent on all services
- Check whether your DNS change is already live for users of those services
- Have an automated check which tells you when something is not right

You can see the current table of nameservers oncluded on the build by browsing the [nameservers.yaml](nameservers.yaml) file.

## Usage

```bash
# ./dns_check --help
Usage of ./dns_check:
  -a, --assert=[]: Exit with exit code 2 when these DNS entries were not found
      --assert-threshold=100: If used with -a fail when not at least N percent of the nameservers had the expected result
  -f, --full-scan[=false]: Scan all nameservers included in this build
  -q, --quiet[=false]: Do not communicate by text, use only exit codes
  -s, --short[=true]: Use short notation (only when using assertion)
      --version[=false]: Print version and exit
```

Use case: I know the IP of my domain and want to check whether all services report that IP

```bash
# ./dns_check -a "188.40.126.69" A luzifer.io
[Level3] (209.244.0.3:53) ✓
[Level3] (209.244.0.4:53) ✓
[Verisign] (64.6.64.6:53) ✓
[Verisign] (64.6.65.6:53) ✓
[Google] (8.8.8.8:53) ✓
[Google] (8.8.4.4:53) ✓
[OpenDNS Home] (208.67.222.222:53) ✓
[OpenDNS Home] (208.67.220.220:53) ✓
```

Use case: Just tell me the IP of any domain

```bash
# ./dns_check A luzifer.io
[Google] (8.8.8.8:53)
 - 188.40.126.69
[Google] (8.8.4.4:53)
 - 188.40.126.69
[Level3] (209.244.0.3:53)
 - 188.40.126.69
[Level3] (209.244.0.4:53)
 - 188.40.126.69
[Verisign] (64.6.64.6:53)
 - 188.40.126.69
[Verisign] (64.6.65.6:53)
 - 188.40.126.69
[OpenDNS Home] (208.67.222.222:53)
 - 188.40.126.69
[OpenDNS Home] (208.67.220.220:53)
 - 188.40.126.69
```
