# Setting up GITM
--- 

## Follow a guide for your target device to install the custom CA certificate
---
The first time GITM starts up, a custom CA certificate will be generated, 
and saved to GITMs config directory.  

See Settings for where that is on your machine.

GITM uses this CA certificate to sign the MITM certificates 
used to impersonate the sites that are proxied through GITM.

If you don't configure your target device to trust this CA certificate, 

you will get warnings about self signed certificates being used, and most applications will
refuse to send traffic through the proxy.

## Follow a guide for configuring your device to use a SOCKS5 proxy
---
You can find the socks5 proxy host and port that GITM will use, and you can change it if necessary. 

Make sure the host is reachable from your target device.

