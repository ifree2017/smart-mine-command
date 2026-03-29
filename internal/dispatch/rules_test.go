package dispatch

import (
	"testing"

	"smart-mine-command/internal/model"
)

func TestEvaluate_HighLevel(t *testing.T) {
	alert := model.Alert{ID: "a1", Level: 4, GasType: "T", Location: "001"}
	cmds := Evaluate(alert, DefaultRules)

	// level 4 >= rule {LevelMin:4, Action:"broadcast"}
	found := false
	for _, c := range cmds {
		if c.Type == "broadcast" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should generate broadcast command for level 4")
	}
}

func TestEvaluate_MediumLevel(t *testing.T) {
	alert := model.Alert{ID: "a2", Level: 3, GasType: "CO", Location: "002"}
	cmds := Evaluate(alert, DefaultRules)

	// level 3 >= rule {LevelMin:3, Action:"ws"}
	found := false
	for _, c := range cmds {
		if c.Type == "ws" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should generate ws command for level 3")
	}
}

func TestEvaluate_LowLevel(t *testing.T) {
	alert := model.Alert{ID: "a3", Level: 1, GasType: "T", Location: "003"}
	cmds := Evaluate(alert, DefaultRules)

	// level 1 >= rule {LevelMin:1, Action:"ws"}
	if len(cmds) == 0 {
		t.Error("should generate ws command for level 1")
	}
}

func TestEvaluate_GasTypeFilter(t *testing.T) {
	rules := []Rule{
		{LevelMin: 1, GasType: "T", Action: "ws"},
		{LevelMin: 1, GasType: "CO", Action: "broadcast"},
	}

	alertT := model.Alert{Level: 1, GasType: "T"}
	cmdsT := Evaluate(alertT, rules)
	if len(cmdsT) != 1 || cmdsT[0].Type != "ws" {
		t.Errorf("T gas: got %v", cmdsT)
	}

	alertCO := model.Alert{Level: 1, GasType: "CO"}
	cmdsCO := Evaluate(alertCO, rules)
	if len(cmdsCO) != 1 || cmdsCO[0].Type != "broadcast" {
		t.Errorf("CO gas: got %v", cmdsCO)
	}
}
