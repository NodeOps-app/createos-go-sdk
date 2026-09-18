package structs

import (
	"fmt"
	"strings"
	"testing"
)

func TestSandboxAccessTokenCreateResponseFormattingRedactsToken(t *testing.T) {
	response := SandboxAccessTokenCreateResponse{Token: "skp_sb_secret", Enabled: true}
	for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		output := fmt.Sprintf(format, response)
		if strings.Contains(output, response.Token) || !strings.Contains(output, "[REDACTED]") {
			t.Errorf("format %q exposed token or omitted redaction: %s", format, output)
		}
	}
	if response.Token != "skp_sb_secret" {
		t.Fatal("formatting changed the plaintext token field")
	}
}
