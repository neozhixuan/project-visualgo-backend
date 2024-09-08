package utils

import "strconv"

// Helper function to safely parse price values
func parsePrice(val interface{}) float64 {
	if str, ok := val.(string); ok {
		if price, err := strconv.ParseFloat(str, 64); err == nil {
			return price
		}
	}
	return 0
}
