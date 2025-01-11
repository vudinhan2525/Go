package util

type CurrencyType string

const (
	USD CurrencyType = "USD"
	EUR CurrencyType = "EUR"
	CAD CurrencyType = "CAD"
	VND CurrencyType = "VND"
)

func IsSupportedCurrency(currency string) bool {
	switch CurrencyType(currency) {
	case EUR, USD, CAD, VND:
		return true
	}
	return false
}
