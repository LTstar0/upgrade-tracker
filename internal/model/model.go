package model

import (
	"strings"
	"time"
)

type Client struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Contact        string    `json:"contact"`
	Note           string    `json:"note"`
	CurrentVersion string    `json:"current_version"`
	UpgradeCount   int       `json:"upgrade_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpgradeRecord struct {
	ID          int       `json:"id"`
	ClientID    int       `json:"client_id"`
	Version     string    `json:"version"`
	UpgradeDate string    `json:"upgrade_date"` // "2006-01-02"
	Operator    string    `json:"operator"`
	Tags        []string  `json:"tags"`
	Description string    `json:"description"`
	Files       []string  `json:"files"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductImage struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Type        string    `json:"type"`
	PublicURL   string    `json:"public_url"`
	InternalURL string    `json:"internal_url"`
	ConfigGuide string    `json:"config_guide"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TagsStr converts slice to comma-separated string for DB storage
func TagsStr(tags []string) string { return strings.Join(tags, ",") }
func FilesStr(files []string) string { return strings.Join(files, ",") }

func SplitTrim(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

type Stats struct {
	TotalClients  int `json:"total_clients"`
	TotalUpgrades int `json:"total_upgrades"`
	MonthUpgrades int `json:"month_upgrades"`
}
