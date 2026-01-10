package cli

import (
	"strings"
	"testing"
)

func TestDateFlagsUsageMentionsYMD(t *testing.T) {
	cmd := NewRootCmd()
	for _, name := range []string{"date", "from", "to"} {
		flag := cmd.Flag(name)
		if flag == nil {
			t.Fatalf("missing flag: %s", name)
		}
		if !strings.Contains(flag.Usage, "YYYY-M-D") {
			t.Fatalf("flag %s usage missing YYYY-M-D: %s", name, flag.Usage)
		}
	}
}
