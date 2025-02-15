package config

import "flag"

type Config struct {
    HttpListenUri string
    TlsListenUri string
    Debug bool
}

func ParseFlags () Config {
    conf := Config{}

    flag.StringVar(&conf.HttpListenUri, "u", "127.0.0.1:8080", "HTTP server listen uri")
    flag.StringVar(&conf.TlsListenUri, "us", "127.0.0.1:8443", "HTTPS server listen uri")
    flag.BoolVar(&conf.Debug, "d", false, "Enable debug logging")

    flag.Parse()

    return conf
}

