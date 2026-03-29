package kj90x

import (
	"testing"
)

func TestParse_ValidFrame(t *testing.T) {
	frame := "$KJ90,ALM,001,T,0.5,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if alert.DeviceAddr != "001" {
		t.Errorf("DeviceAddr: got %s, want 001", alert.DeviceAddr)
	}
	if alert.GasType != "T" {
		t.Errorf("GasType: got %s, want T", alert.GasType)
	}
	if alert.Status != "1" {
		t.Errorf("Status: got %s, want 1", alert.Status)
	}
}

func TestParse_InvalidFrame(t *testing.T) {
	_, err := Parse("INVALID")
	if err == nil {
		t.Error("should return error for invalid frame")
	}
}

func TestParse_LevelHigh(t *testing.T) {
	// CO with high concentration should trigger high level
	frame := "$KJ90,ALM,002,CO,50.0,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	modelAlert := alert.ToModel()
	if modelAlert.Level < 3 {
		t.Errorf("Level for high CO: got %d, want >=3", modelAlert.Level)
	}
}

func TestAlert_ToEvent(t *testing.T) {
	a := &Alert{DeviceAddr: "001", GasType: "T", Status: "1"}
	e := a.ToEvent()
	if e.Type != "alert" {
		t.Errorf("Type: got %s, want alert", e.Type)
	}
}

func TestAlert_ToModel_LowConcentration(t *testing.T) {
	// Low T gas concentration should give level 1 or 2
	frame := "$KJ90,ALM,003,T,0.1,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level < 1 || m.Level > 2 {
		t.Errorf("Low T gas level: got %d, want 1 or 2", m.Level)
	}
	if m.Location != "003" {
		t.Errorf("Location: got %s, want 003", m.Location)
	}
}

func TestParse_InsufficientFields(t *testing.T) {
	_, err := Parse("$KJ90,ALM,001,T")
	if err == nil {
		t.Error("should return error for insufficient fields")
	}
}

func TestParse_NotAlertFrame(t *testing.T) {
	_, err := Parse("$KJ90,OTHER,001,T,0.5,20260329120000,1*XX")
	if err == nil {
		t.Error("should return error for non-alert frame")
	}
}

func TestAlert_ToModel_EmptyDeviceAddr(t *testing.T) {
	frame := "$KJ90,ALM,,T,0.5,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Location != "未知位置" {
		t.Errorf("Empty DeviceAddr: got %s, want 未知位置", m.Location)
	}
}
