package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteVersionJSONIncludesProfile(t *testing.T) {
	oldProfile := runtimeProfile
	runtimeProfile = profileExperimental
	defer func() { runtimeProfile = oldProfile }()

	var buf bytes.Buffer
	if err := writeVersion(&buf, true); err != nil {
		t.Fatalf("writeVersion() error = %v", err)
	}

	var info VersionInfo
	if err := json.Unmarshal(buf.Bytes(), &info); err != nil {
		t.Fatalf("version JSON is invalid: %v", err)
	}
	if info.Profile != profileExperimental {
		t.Fatalf("Profile = %q, want %q", info.Profile, profileExperimental)
	}
	if info.GoVersion == "" || info.OS == "" || info.Arch == "" {
		t.Fatalf("runtime fields should be populated: %+v", info)
	}
}

func TestValidateRuntimeProfileDoesNotGateCommands(t *testing.T) {
	oldProfile := runtimeProfile
	runtimeProfile = profileStable
	defer func() { runtimeProfile = oldProfile }()

	if err := validateRuntimeProfile(payloadCmd, nil); err != nil {
		t.Fatalf("stable profile should not block existing commands: %v", err)
	}
}

func TestValidateRuntimeProfileRejectsUnknownProfile(t *testing.T) {
	oldProfile := runtimeProfile
	runtimeProfile = "unknown"
	defer func() { runtimeProfile = oldProfile }()

	if err := validateRuntimeProfile(rootCmd, nil); err == nil {
		t.Fatal("validateRuntimeProfile() error = nil, want error")
	}
}

func TestDoctorReportCommandInventory(t *testing.T) {
	report := runDoctorReport(true)
	if report.Summary.Total == 0 {
		t.Fatal("doctor report should include checks")
	}
	if len(report.Commands) == 0 {
		t.Fatal("doctor report should include command inventory when requested")
	}

	var foundPayload bool
	for _, row := range report.Commands {
		if row.Name == "payload" {
			foundPayload = true
			if row.Class != classExperimental {
				t.Fatalf("payload class = %q, want %q", row.Class, classExperimental)
			}
		}
	}
	if !foundPayload {
		t.Fatal("command inventory should include payload")
	}
}

func TestVerifyReportText(t *testing.T) {
	report := verifyReport{
		Profile: profileStable,
		Checks:  []doctorCheck{{Name: "sample", Status: statusOK, Message: "ready"}},
		Summary: checkSummary{
			OK:    1,
			Total: 1,
		},
	}

	var buf bytes.Buffer
	if err := writeVerifyReport(&buf, report, false); err != nil {
		t.Fatalf("writeVerifyReport() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "GoCat verify") || !strings.Contains(out, "[OK] sample") {
		t.Fatalf("verify text output missing expected content: %q", out)
	}
}
