# vpn_proxy
simple relible proxy server for webscraping

# setup
the proxies come preset with openvpn configs assuming you are using nordvpn (if this is not the vpn you are using simply reset the configs based on the instructions below)

"vpn-auth.txt" 
This file should contain your VPN username on the first line and your VPN password on the second line
or if you dont have it

go build main.go vpn.go

sudo main #this is needed for ovpn. 

## setting your own ovpn configs
```bash
get_proxy_files.go 
```

this would setup all the vpn configs to the proxies

# TCP considerations

I was trying to disable having the keepalives stay open but honestly there was no clean way to do it. 
because of this we get this fairly anoying situation where it takes WAY too long for it to close (multiple seconds)

the only safe way I can think of to handle it is to manualy kill the proxy out right. this is not really a pattern I like so we will just have to take long die times