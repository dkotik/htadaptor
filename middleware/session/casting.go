package session

import "strconv"

func (c *sessionContext) Int(key string) int {
	// type switch is really important
	switch v := c.values[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case uint:
		return int(v)
	case uint64:
		return int(v)
	case string:
		cast, _ := strconv.Atoi(v)
		return cast
	default:
		return 0
	}
}

func (c *sessionContext) Int64(key string) int64 {
	// type switch is really important
	switch v := c.values[key].(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case float32:
		return int64(v)
	case uint:
		return int64(v)
	case uint64:
		return int64(v)
	case string:
		cast, _ := strconv.Atoi(v)
		return int64(cast)
	default:
		return 0
	}
}

func (c *sessionContext) Float32(key string) float32 {
	// type switch is really important
	switch v := c.values[key].(type) {
	case float32:
		return v
	case int:
		return float32(v)
	case float64:
		return float32(v)
	case int64:
		return float32(v)
	case uint:
		return float32(v)
	case uint64:
		return float32(v)
	case string:
		cast, _ := strconv.ParseFloat(v, 64)
		return float32(cast)
	default:
		return 0
	}
}

func (c *sessionContext) Float64(key string) float64 {
	// type switch is really important
	switch v := c.values[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case string:
		cast, _ := strconv.ParseFloat(v, 64)
		return cast
	default:
		return 0
	}
}
