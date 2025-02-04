package clients

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	/* not using this package directly, but the import
	will register the sqlite driver for 'database/sql' */
	_ "github.com/mattn/go-sqlite3"
)

type FirefoxClient struct {
	tokens            map[string]string
	tokenCreationTime int
	target            *url.URL
}

func getTemporaryDatabasePath() string {
	temp, err := os.CreateTemp("", "stravaheatmapproxycookies*.sqlite")
	if err != nil {
		log.Fatal("Failed to create temp file: ", err)
	}
	temp.Close()
	return temp.Name()
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

func createCookieDatabaseBackup(path string) (string, error) {
	suffixes := []string{"", "-wal", "-shm"}
	temp := getTemporaryDatabasePath()
	for _, suffix := range suffixes {
		if err := copyFile(path+suffix, temp+suffix); err != nil {
			return temp, err
		}
	}
	return temp, nil
}

func cleanupCookieDatabaseBackup(path string) error {
	suffixes := []string{"", "-wal", "-shm"}
	for _, suffix := range suffixes {
		err := os.Remove(path + suffix)
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (client *FirefoxClient) readCookieDatabase(path string) error {
	backupPath, err := createCookieDatabaseBackup(path)
	if err != nil {
		log.Fatal("Failed to copy database: ", err)
	}
	defer cleanupCookieDatabaseBackup(backupPath)

	db, err := sql.Open("sqlite3", "file:"+backupPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// commit pending changes from wal file to database
	db.Exec("PRAGMA wal_checkpoint(FULL);")

	query := "SELECT name, value, creationTime FROM moz_cookies WHERE name IN (?, ?, ?) AND host = '.strava.com'"
	rows, err := db.Query(query, "CloudFront-Policy", "CloudFront-Key-Pair-Id", "CloudFront-Signature")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, value string
		var creationTime int
		if err := rows.Scan(&name, &value, &creationTime); err != nil {
			log.Fatal(err)
		}
		client.tokens[name] = value
		client.tokenCreationTime = creationTime
	}
	return nil
}

func (client *FirefoxClient) isNewerThan(other *FirefoxClient) bool {
	return client.tokenCreationTime > other.tokenCreationTime
}

func findCookieDatabases(profilesDir string) ([]string, error) {
	var cookieDatabases []string
	err := filepath.Walk(profilesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			cookiePath := filepath.Join(path, "cookies.sqlite")
			if _, err := os.Stat(cookiePath); err == nil {
				// if 'cookies.sqlite' exists in this profile directory
				cookieDatabases = append(cookieDatabases, cookiePath)
			}
		}
		return nil
	})
	return cookieDatabases, err
}

func getProfilesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var profilesDir string
	switch runtime.GOOS {
	case "linux":
		// Linux: ~/.mozilla/firefox/
		profilesDir = filepath.Join(homeDir, ".mozilla", "firefox")
	case "darwin":
		// macOS: ~/Library/Application Support/Firefox/
		profilesDir = filepath.Join(homeDir, "Library", "Application Support", "Firefox")
	case "windows":
		// Windows: C:\Users\<username>\AppData\Roaming\Mozilla\Firefox\
		profilesDir = filepath.Join(homeDir, "AppData", "Roaming", "Mozilla", "Firefox")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return profilesDir, nil
}

func (client *FirefoxClient) isValid() bool {
	return client.tokenCreationTime > 0 && len(client.tokens) == 3
}

func NewFirefoxClient(target *url.URL) (*FirefoxClient, error) {
	var client *FirefoxClient
	profilesDir, err := getProfilesDir()
	if err != nil {
		return client, err
	}
	cookieDatabases, err := findCookieDatabases(profilesDir)
	if err != nil {
		return client, err
	}
	for _, databasePath := range cookieDatabases {
		c := &FirefoxClient{}
		c.target = target
		c.tokens = make(map[string]string)
		if err := c.readCookieDatabase(databasePath); err != nil {
			return client, err
		}
		if !c.isValid() {
			continue
		}
		if client == nil {
			client = c
		} else if c.isNewerThan(client) {
			client = c
		}
	}
	return client, nil
}

func (client *FirefoxClient) GetCloudFrontTokens() map[string]string {
	return client.tokens
}

func (client *FirefoxClient) GetCookies(url *url.URL) []*http.Cookie {
	var cookies []*http.Cookie
	for name, value := range client.tokens {
		cookies = append(cookies, &http.Cookie{
			Name:  name,
			Value: value,
		})
	}
	return cookies
}

func (client *FirefoxClient) GetTarget() *url.URL {
	return client.target
}
