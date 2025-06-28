# Firefox setup

To setup firefox for use with gitm, you will need to install an extension like [FoxyProxy](https://getfoxyproxy.org/).

While other extensions will work, this tutorial will use FoxyProxy as an example.

1. Locate gitm sock5 proxy url
Open GITM, then navigate to File > Settings:

![](docs:firefox-1.png)

Next, locate the socks5 proxy url. Update it to a port that firefox will be able to access.
(If firefox will be running on the same computer as gitm, localhost will work fine.

![](docs:firefox-2.png)

Finally using the above socks5 url, configure a proxy in Foxyproxy for gitm:

![](docs:firefox-3.png)

Now, you can set foxyproxy to use the configured proxy, and you should be able to see

your web traffic in gitm:

![](docs:firefox-4.png)

![](docs:firefox-5.png)
![](docs:firefox-6.png)

Well, with an annoying warning about unknown issuer... while this may work for 

simple intercepts, we want more!!

2. Adding GITM ca certificate to firefox

