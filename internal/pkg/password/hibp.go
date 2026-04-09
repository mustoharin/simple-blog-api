package password

import (
	"bufio"
	"crypto/sha1" // #nosec G401 - SHA-1 required by HIBP k-anonymity API
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultHIBPBaseURL = "https://api.pwnedpasswords.com/range/"

// HIBPConfig allows overriding the HIBP API URL (used in tests).
type HIBPConfig struct {
	BaseURL string
}

type hibpChecker struct {
	client  *http.Client
	baseURL string
}

func newHIBPChecker(cfg *HIBPConfig) *hibpChecker {
	baseURL := defaultHIBPBaseURL
	if cfg != nil && cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}
	return &hibpChecker{
		client:  &http.Client{Timeout: 3 * time.Second},
		baseURL: baseURL,
	}
}

// isPwned checks if the given plain password appears in HIBP using k-anonymity.
// Returns false (fail-open) on any network or parsing error.
func (h *hibpChecker) isPwned(plain string) bool {
	// #nosec G401 - SHA-1 required by the HIBP k-anonymity API
	sum := sha1.Sum([]byte(plain))
	hash := strings.ToUpper(fmt.Sprintf("%x", sum))
	prefix := hash[:5]
	suffix := hash[5:]

	resp, err := h.client.Get(h.baseURL + prefix) // #nosec G107
	if err != nil {
		return false // fail-open
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false // fail-open
	}

	scanner := bufio.NewScanner(io.LimitReader(resp.Body, 1<<20)) // 1MB cap
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(parts[0], suffix) && parts[1] != "0" {
			return true
		}
	}
	return false
}
