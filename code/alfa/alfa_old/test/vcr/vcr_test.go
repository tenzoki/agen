package vcr_test

import (
    "os"
    "strings"
    "testing"

    "alfa/internal/vcr"
)

func TestVcrBasic(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(tmp+"/f1.txt", []byte("init"), 0644)

	// Init Vcr
	v := vcr.NewVcr("tester", tmp)

	// Change f1.txt, then commit changes
	os.WriteFile(tmp+"/f1.txt", []byte("first"), 0644)
	commit1 := v.Commit("initial commit")
	if commit1 == "" || commit1 == "?" {
		t.Error("Commit should return hash")
	}

	// Change and commit again
	os.WriteFile(tmp+"/f1.txt", []byte("second"), 0644)
	commit2 := v.Commit("second commit")
	if commit2 == "" || commit2 == "?" {
		t.Error("Second commit should return hash")
	}
	if commit1 == commit2 {
		t.Error("Commit hashes should be different")
	}

	// Get history
	hist := v.GetHistory()
	joined := strings.Join(hist, ",")
	if !strings.Contains(joined, "initial commit") || !strings.Contains(joined, "second commit") {
		t.Errorf("Unexpected history: %v", hist)
	}

	// Purge should remove .git
	if err := v.Purge(); err != nil {
		t.Errorf("Purge error: %v", err)
	}
	if _, err := os.Stat(tmp + "/.git"); !os.IsNotExist(err) {
		t.Error(".git folder not removed after Purge")
	}
}
