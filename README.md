[![Go Reference](https://pkg.go.dev/badge/github.com/redawl/gitm.svg)](https://pkg.go.dev/github.com/redawl/gitm)
![CI](https://github.com/redawl/gitm/actions/workflows/build.yaml/badge.svg)
![go report card](https://goreportcard.com/badge/github.com/redawl/gitm)
# Gopher in the middle


![Gopher](assets/Icon.png)

GITM is a man in the middle proxy that allows inspecting tls encrypted https data. 
It is currently in heavy development. More info to come!

# Screenshot
![Packet capture](assets/screenshot-darkmode.png)

# Features
- Intercept http and https requests and responses between a client you control, and any server
- Support for intercepting websocket traffic
- Automatically uncompresses many compression types, such as gzip and deflate.
- Decode parts of intercepted packets. Ex: Hex, Base64, urlencoding, etc.
- Save intercepted packets for later analysis, using open humanreadable format (yes, json lol)
- Add your own decoding mappings

# Installation
If you have go installed, you can grab the latest version of the package:
```bash
go install github.com/redawl/gitm@latest
```

Or, you can download precompiled binaries from the releases page:
https://github.com/redawl/gitm/releases
