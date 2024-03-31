# vpn_proxy
simple relible proxy server for webscraping

# setup
the proxies come preset with openvpn configs assuming you are using nordvpn (if this is not the vpn you are using simply reset the configs based on the instructions below)

"vpn-auth.txt" 
This file should contain your VPN username on the first line and your VPN password on the second line
or if you dont have it

## setting your own ovpn configs
```bash
get_proxy_files.go 
```

this would setup all the vpn configs to the proxies