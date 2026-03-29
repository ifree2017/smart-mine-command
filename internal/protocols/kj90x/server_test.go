package kj90x

import (
	"testing"
	"time"

	"smart-mine-command/internal/eventbus"
)

func TestNewServer(t *testing.T) {
	eb := eventbus.New()
	s := NewServer(":54321", eb)
	if s.addr != ":54321" {
		t.Errorf("addr: got %s, want :54321", s.addr)
	}
	if s.eb != eb {
		t.Error("eb not set correctly")
	}
}

func TestServer_Stop_WithoutRun(t *testing.T) {
	eb := eventbus.New()
	s := NewServer(":54321", eb)
	// Calling Stop before Run should not panic
	s.Stop()
}

func TestServer_RunAndStop(t *testing.T) {
	eb := eventbus.New()
	s := NewServer("localhost:0", eb)
	err := s.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}



func TestParse_TGas_Level4(t *testing.T) {
	// T gas >= 1.5 and < 2.0 => level 4
	frame := "$KJ90,ALM,010,T,1.5,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level != 4 {
		t.Errorf("Level for T=1.5: got %d, want 4", m.Level)
	}
}

func TestParse_TGas_Level3(t *testing.T) {
	// T gas >= 1.0 and < 1.5 => level 3
	frame := "$KJ90,ALM,010,T,1.0,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level != 3 {
		t.Errorf("Level for T=1.0: got %d, want 3", m.Level)
	}
}

func TestParse_COGas_Level2(t *testing.T) {
	// CO >= 10 and < 24 => level 2
	frame := "$KJ90,ALM,010,CO,10.0,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level != 2 {
		t.Errorf("Level for CO=10: got %d, want 2", m.Level)
	}
}

func TestParse_COGas_Level3(t *testing.T) {
	// CO >= 24 and < 50 => level 3
	frame := "$KJ90,ALM,010,CO,24.0,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level != 3 {
		t.Errorf("Level for CO=24: got %d, want 3", m.Level)
	}
}

func TestParse_COGas_Level4(t *testing.T) {
	// CO >= 50 and < 100 => level 4
	frame := "$KJ90,ALM,010,CO,50.0,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level != 4 {
		t.Errorf("Level for CO=50: got %d, want 4", m.Level)
	}
}

func TestParse_COGas_Level5(t *testing.T) {
	// CO >= 100 => level 5
	frame := "$KJ90,ALM,010,CO,100.0,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level != 5 {
		t.Errorf("Level for CO=100: got %d, want 5", m.Level)
	}
}

func TestParse_TGas_Level5(t *testing.T) {
	// T >= 2.0 => level 5
	frame := "$KJ90,ALM,010,T,2.0,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level != 5 {
		t.Errorf("Level for T=2.0: got %d, want 5", m.Level)
	}
}

func TestParse_UnknownGas(t *testing.T) {
	// Unknown gas type should default to level 1
	frame := "$KJ90,ALM,010,UNKNOWN,50.0,20260329120000,1*XX"
	alert, err := Parse(frame)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	m := alert.ToModel()
	if m.Level != 1 {
		t.Errorf("Level for unknown gas: got %d, want 1", m.Level)
	}
}
