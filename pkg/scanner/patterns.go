package scanner

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	gitleaksconfig "github.com/zricethezav/gitleaks/v8/config"

	"github.com/leaktk/scanner/pkg/config"
)

// Patterns acts as an abstraction for fetching different scanner patterns
// and keeping them up to date and cached
type Patterns struct {
	config *config.Patterns
	client HTTPClient
}

// NewPatterns returns a configured instance of Patterns
func NewPatterns(cfg *config.Patterns, client HTTPClient) *Patterns {
	return &Patterns{
		config: cfg,
		client: client,
	}
}

func (p *Patterns) fetchGitleaksConfig() (string, error) {
	url, err := url.JoinPath(
		p.config.Server.URL, "patterns", "gitleaks", p.config.Gitleaks.Version,
	)

	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if len(p.config.Server.AuthToken) > 0 {
		request.Header.Add(
			"Authorization",
			fmt.Sprintf("Bearer %s", p.config.Server.AuthToken),
		)
	}

	response, err := p.client.Do(request)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	return string(body), err
}

func (p *Patterns) parseGitleaksConfig(rawConfig string) (*gitleaksconfig.Config, error) {
	// From https://github.com/gitleaks/gitleaks/blob/79cac73f7267f4a48f4bc73db11e105a6098a836/cmd/root.go#L123
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(strings.NewReader(rawConfig)); err != nil {
		return nil, err
	}

	// From https://github.com/gitleaks/gitleaks/blob/79cac73f7267f4a48f4bc73db11e105a6098a836/cmd/root.go#L160
	var vc gitleaksconfig.ViperConfig
	if err := viper.Unmarshal(&vc); err != nil {
		return nil, err
	}

	cfg, err := vc.Translate()
	if err != nil {
		return nil, err
	}

	cfg.Path = p.config.Gitleaks.ConfigPath

	return &cfg, nil
}

// Gitleaks returns a Gitleaks config object if it's able to
// TODO: make sure this is safe for concurrency
func (p *Patterns) Gitleaks() (*gitleaksconfig.Config, error) {
	if p.config.Gitleaks.Config == nil {
		// TODO: load patterns from FS if they exist and are newer than the refresh time

		if !p.config.Autofetch {
			return p.config.Gitleaks.Config, fmt.Errorf("could not autofetch gitleaks config because autofetch is disabled")
		}

		rawConfig, err := p.fetchGitleaksConfig()
		if err != nil {
			return p.config.Gitleaks.Config, err
		}

		p.config.Gitleaks.Config, err = p.parseGitleaksConfig(rawConfig)
		if err != nil {
			return p.config.Gitleaks.Config, err
		}

		err = os.MkdirAll(filepath.Dir(p.config.Gitleaks.ConfigPath), 0700)
		if err != nil {
			return p.config.Gitleaks.Config, err
		}

		configFile, err := os.Create(p.config.Gitleaks.ConfigPath)
		if err != nil {
			return p.config.Gitleaks.Config, err
		}
		defer configFile.Close()

		_, err = configFile.WriteString(rawConfig)
		if err != nil {
			return p.config.Gitleaks.Config, err
		}
	}

	return p.config.Gitleaks.Config, nil
}
