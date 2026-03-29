package kj90x

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
)

type Alert struct {
	DeviceAddr string
	GasType    string
	Concentration float64
	ReadTime   string
	Status     string
}

func (a *Alert) ToEvent() eventbus.Event {
	return eventbus.Event{
		Type:   eventbus.EventTypeAlert,
		Source: "kj90x",
		Data: map[string]interface{}{
			"device_addr":  a.DeviceAddr,
			"gas_type":     a.GasType,
			"concentration": a.Concentration,
			"read_time":    a.ReadTime,
			"status":       a.Status,
		},
	}
}

func (a *Alert) ToModel() model.Alert {
	level := 1
	switch a.GasType {
	case "T": // 甲烷/瓦斯
		if a.Concentration >= 2.0 {
			level = 5
		} else if a.Concentration >= 1.5 {
			level = 4
		} else if a.Concentration >= 1.0 {
			level = 3
		} else if a.Concentration >= 0.5 {
			level = 2
		}
	case "CO": // 一氧化碳
		if a.Concentration >= 100 {
			level = 5
		} else if a.Concentration >= 50 {
			level = 4
		} else if a.Concentration >= 24 {
			level = 3
		} else if a.Concentration >= 10 {
			level = 2
		}
	}
	if level == 0 {
		level = 1
	}

	loc := a.DeviceAddr
	if loc == "" {
		loc = "未知位置"
	}
	t, _ := time.Parse("20060102150405", a.ReadTime)
	if t.Year() == 0 {
		t = time.Now()
	}

	return model.Alert{
		ID:        fmt.Sprintf("KJ90-%d", time.Now().UnixNano()),
		Source:    "kj90x",
		Level:     level,
		Location:  loc,
		GasType:   a.GasType,
		Value:     a.Concentration,
		Status:    "open",
		CreatedAt: t,
	}
}

// Parse 解析 kj90x 数据帧
// 格式：$KJ90,ALM,设备地址,气体类型,浓度,时间,状态*XX
func Parse(line string) (*Alert, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "$KJ90") {
		return nil, fmt.Errorf("invalid frame prefix")
	}

	// 去掉 $KJ90 前缀和末尾校验
	rest := strings.TrimPrefix(line, "$KJ90")
	rest = strings.TrimPrefix(rest, ",")

	// 去掉末尾 *XX
	if idx := strings.Index(rest, "*"); idx >= 0 {
		rest = rest[:idx]
	}

	parts := strings.Split(rest, ",")
	if len(parts) < 6 {
		return nil, fmt.Errorf("insufficient fields: %d", len(parts))
	}

	if parts[0] != "ALM" {
		return nil, fmt.Errorf("not an alert frame: %s", parts[0])
	}

	conc, _ := strconv.ParseFloat(parts[3], 64)

	return &Alert{
		DeviceAddr:    parts[1],
		GasType:       parts[2],
		Concentration: conc,
		ReadTime:      parts[4],
		Status:        parts[5],
	}, nil
}
