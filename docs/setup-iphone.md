## Setting up gitm on iPhone

1. Install custom CA certificate

GITM contains a webserver that will serve the custom root certificate. We can use this to establish trust between the iPhone and GITM


First, check what the cacert proxy url is in GITM. Make sure it is set to use the ip address on your local network (not localhost or 127.0.0.1):

![](docs:Iphone-setup1.png)


Then, navigate to "http://${CACERT_PROXY_URL}/ca.crt" in Safari. You should see the below popup:

![](docs:Iphone-setup2.png)

Select allow, and open the iPhone settings app. You should see a new "Profile downloaded" option near the top of the app:

![](docs:Iphone-setup3.png)

Install the profile:

![](docs:Iphone-setup4.png)
![](docs:Iphone-setup5.png)

Verify you see the GITM Configuration profile under "VPN Device and management":

![](docs:Iphone-setup6.png)

Under certificate trust settings, enable the GITM Inc Root certificate. 

2. Configure iPhone to use the socks5 proxy

The same proxy server we used in step one also servers a proxy.pac file, which tells the iPhone where the reverse proxy is.

First navigate to the settings page for the wifi network the iPhone is connected to:

