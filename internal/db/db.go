package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/redawl/gitm/internal/util"
)

type DomainInfo struct {
	Domain  string
	Cert    []byte
	PrivKey []byte
}

func getConn() (*sql.DB, error) {
	configDir, err := util.GetConfigDir()

	if err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite3", configDir+"/domains.db")

	if err != nil {
		return nil, err
	}

	if _, err := conn.Exec(`
        CREATE TABLE IF NOT EXISTS DOMAINS (
            domain varchar(100) PRIMARY KEY,
            cert BLOB,
            privkey BLOB
        )
    `); err != nil {
		return nil, err
	}

	return conn, nil
}

func GetDomains() ([]DomainInfo, error) {
	conn, err := getConn()

	if err != nil {
		return nil, err
	}

	rows, err := conn.Query("SELECT domain, cert, privkey from DOMAINS")

	if err != nil {
		return nil, err
	}

	domainInfos := make([]DomainInfo, 0)
	for rows.Next() {
		domainInfo := DomainInfo{}
		if err := rows.Scan(&domainInfo.Domain, &domainInfo.Cert, &domainInfo.PrivKey); err != nil {
			return nil, err
		}

		domainInfos = append(domainInfos, domainInfo)
	}

	return domainInfos, nil
}

func GetDomain(domain string) (*DomainInfo, error) {
	conn, err := getConn()

	if err != nil {
		return nil, err
	}

	row := conn.QueryRow("SELECT domain, cert, privkey from DOMAINS where domain = $1", domain)

	domainInfo := DomainInfo{}
	if err := row.Scan(&domainInfo.Domain, &domainInfo.Cert, &domainInfo.PrivKey); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &domainInfo, nil
}

func AddDomain(domain string, cert []byte, privkey []byte) error {
	conn, err := getConn()

	if err != nil {
		return err
	}

	if _, err := conn.Exec(`
        INSERT INTO DOMAINS (domain, cert, privkey) 
        VALUES ($1, $2, $3) ON CONFLICT (domain) DO NOTHING
    `, domain, cert, privkey); err != nil {
		return err
	}

	return nil
}
