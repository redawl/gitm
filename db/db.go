package db

import (
	"database/sql"

	"com.github.redawl.mitmproxy/util"
	_ "github.com/mattn/go-sqlite3"
)

func getConn() (*sql.DB, error) {
    configDir, err := util.GetConfigDir()

    if err != nil {
        return nil, err
    }

    conn, err := sql.Open("sqlite3",  configDir + "/domains.db")

    if err != nil {
        return nil, err
    }

    if _, err := conn.Exec("CREATE TABLE IF NOT EXISTS DOMAINS (domain varchar(100) PRIMARY KEY)"); err != nil {
        return nil, err
    }

    return conn, nil
}

func GetDomains() ([]string, error) {
    conn, err := getConn()

    if err != nil {
        return nil, err
    }

    rows, err := conn.Query("SELECT domain from DOMAINS")
    
    if err != nil {
        return nil, err
    }

    domains := make([]string, 0)
    for rows.Next() {
        domain := ""
        err := rows.Scan(&domain)
        if err != nil {
            return nil, err
        }

        domains = append(domains, domain)
    }

    return domains, nil
}

func AddDomain(domain string) error {
    conn, err := getConn()

    if err != nil {
        return err
    }

    if _, err := conn.Exec("INSERT INTO DOMAINS (domain) VALUES ($1) ON CONFLICT (domain) DO NOTHING", domain); err != nil {
        return err
    }

    return nil
}
