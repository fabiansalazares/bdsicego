package download

import (
	"testing"

	"github.com/fabiansalazares/bdsicego/internal/config"
)

func TestUpdate(t *testing.T) {
	configuration, err := config.GetConfig()
	if err != nil {
		t.Errorf("TestUpdate: %s", err.Error())
	}

	extractedFiles, err := Update(configuration, true)

	if (extractedFiles == nil || len(extractedFiles) == 0) && (err == nil) {
		t.Fatalf("Update() failed: returned an empty slice, meaning no files were eextracted. No error was reported.")
	}

	if err != nil {
		t.Fatalf("Update() returned an error: %s\n", err.Error())
	}

	// we should also check that the .xer files with the updated series have actually been copied correctly
	// how to do that? calculate some hash files? or just rely on Update() return?
}

func TestDownloadFullDatabase(t *testing.T) {
	configuration, err := config.GetConfig()
	if err != nil {
		t.Errorf("TestDownloadFullDatabase: %s", err.Error())
	}

	extractedFiles, err := DownloadFullDatabase(configuration, true)

	if (extractedFiles == nil || len(extractedFiles) == 0) && err == nil {
		t.Fatalf("DownloadFullDatabase() failed: returned an empty slice, meaning no files were eextracted. No error was reported.")
	}

	if err != nil {
		t.Fatalf("DownloadFullDatabase() returned an error: %s\n", err.Error())
	}
}

func TestBulletin(t *testing.T) {
	configuration, err := config.GetConfig()
	if err != nil {
		t.Errorf("TestBulletin: %s", err.Error())
	}

	err = Bulletin(configuration)

	if err != nil {
		t.Fatalf("Bulletin() returned an error: %s\n", err.Error())
	}
}
