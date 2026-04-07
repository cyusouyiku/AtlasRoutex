package valueobject

import "fmt"

// Budget 预算值对象,主要包含货币的定义，货币汇率转换，数学运算，比较操作，还有BugdgetRange
type Currency string

const (
	CurrencyUSD   Currency = "USD"   // United States Dollar
	CurrencyCNY   Currency = "CNY"   // Chinese Yuan
	CurrencyEUR   Currency = "EUR"   // Euro
	CurrencyJPY   Currency = "JPY"   // Japanese Yen
	CurrencyGBP   Currency = "GBP"   // British Pound
	CurrencyAUD   Currency = "AUD"   // Australian Dollar
	CurrencyCAD   Currency = "CAD"   // Canadian Dollar
	CurrencyCHF   Currency = "CHF"   // Swiss Franc
	CurrencyHKD   Currency = "HKD"   // Hong Kong Dollar
	CurrencySGD   Currency = "SGD"   // Singapore Dollar
	CurrencyOther Currency = "OTHER" // Other
)

// 这里其实就是简单写个汇率，没有调用api，后续可以改进提升一下
func transformCurrencyFromUSD(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyUSD {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	} else if from == CurrencyUSD {
		switch to {
		case CurrencyCNY:
			return amount * 7.0, nil
		case CurrencyEUR:
			return amount * 0.85, nil
		case CurrencyJPY:
			return amount * 110.0, nil
		case CurrencyGBP:
			return amount * 0.75, nil
		case CurrencyAUD:
			return amount * 1.35, nil
		case CurrencyCAD:
			return amount * 1.25, nil
		case CurrencyCHF:
			return amount * 0.92, nil
		case CurrencyHKD:
			return amount * 7.8, nil
		case CurrencySGD:
			return amount * 1.35, nil
		default:
			return 0, fmt.Errorf("不支持此货币: %s", to)
		}
	}
	return 0, fmt.Errorf("不支持此货币: %s", from)
}

func transformCurrencyFromCNY(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyCNY {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	} else if from == CurrencyCNY {
		switch to {
		case CurrencyUSD:
			return amount * 0.14, nil
		case CurrencyEUR:
			return amount * 0.12, nil
		case CurrencyJPY:
			return amount * 15.7, nil
		case CurrencyGBP:
			return amount * 0.11, nil
		case CurrencyAUD:
			return amount * 0.19, nil
		case CurrencyCAD:
			return amount * 0.16, nil
		case CurrencyCHF:
			return amount * 0.13, nil
		case CurrencyHKD:
			return amount * 1.12, nil
		case CurrencySGD:
			return amount * 0.19, nil
		default:
			return 0, fmt.Errorf("不支持此货币: %s", to)
		}
	}
	return 0, fmt.Errorf("不支持此货币: %s", from)
}

func transformCurrencyFromEUR(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyEUR {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	}
	switch to {
	case CurrencyUSD:
		return amount * 1.176, nil
	case CurrencyCNY:
		return amount * 8.235, nil
	case CurrencyJPY:
		return amount * 129.41, nil
	case CurrencyGBP:
		return amount * 0.882, nil
	case CurrencyAUD:
		return amount * 1.588, nil
	case CurrencyCAD:
		return amount * 1.471, nil
	case CurrencyCHF:
		return amount * 1.082, nil
	case CurrencyHKD:
		return amount * 9.176, nil
	case CurrencySGD:
		return amount * 1.588, nil
	default:
		return 0, fmt.Errorf("不支持此货币: %s", to)
	}
}

func transformCurrencyFromJPY(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyJPY {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	}
	switch to {
	case CurrencyUSD:
		return amount * 0.00909, nil
	case CurrencyCNY:
		return amount * 0.0636, nil
	case CurrencyEUR:
		return amount * 0.00773, nil
	case CurrencyGBP:
		return amount * 0.00682, nil
	case CurrencyAUD:
		return amount * 0.01227, nil
	case CurrencyCAD:
		return amount * 0.01136, nil
	case CurrencyCHF:
		return amount * 0.00836, nil
	case CurrencyHKD:
		return amount * 0.0709, nil
	case CurrencySGD:
		return amount * 0.01227, nil
	default:
		return 0, fmt.Errorf("不支持此货币: %s", to)
	}
}

func transformCurrencyFromGBP(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyGBP {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	}
	switch to {
	case CurrencyUSD:
		return amount * 1.333, nil
	case CurrencyCNY:
		return amount * 9.333, nil
	case CurrencyEUR:
		return amount * 1.136, nil
	case CurrencyJPY:
		return amount * 146.67, nil
	case CurrencyAUD:
		return amount * 1.8, nil
	case CurrencyCAD:
		return amount * 1.667, nil
	case CurrencyCHF:
		return amount * 1.227, nil
	case CurrencyHKD:
		return amount * 10.4, nil
	case CurrencySGD:
		return amount * 1.8, nil
	default:
		return 0, fmt.Errorf("不支持此货币: %s", to)
	}
}

func transformCurrencyFromAUD(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyAUD {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	}
	switch to {
	case CurrencyUSD:
		return amount * 0.741, nil
	case CurrencyCNY:
		return amount * 5.185, nil
	case CurrencyEUR:
		return amount * 0.63, nil
	case CurrencyJPY:
		return amount * 81.48, nil
	case CurrencyGBP:
		return amount * 0.556, nil
	case CurrencyCAD:
		return amount * 0.926, nil
	case CurrencyCHF:
		return amount * 0.681, nil
	case CurrencyHKD:
		return amount * 5.778, nil
	case CurrencySGD:
		return amount * 1.0, nil
	default:
		return 0, fmt.Errorf("不支持此货币: %s", to)
	}
}

func transformCurrencyFromCAD(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyCAD {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	}
	switch to {
	case CurrencyUSD:
		return amount * 0.8, nil
	case CurrencyCNY:
		return amount * 5.6, nil
	case CurrencyEUR:
		return amount * 0.68, nil
	case CurrencyJPY:
		return amount * 88.0, nil
	case CurrencyGBP:
		return amount * 0.6, nil
	case CurrencyAUD:
		return amount * 1.08, nil
	case CurrencyCHF:
		return amount * 0.736, nil
	case CurrencyHKD:
		return amount * 6.24, nil
	case CurrencySGD:
		return amount * 1.08, nil
	default:
		return 0, fmt.Errorf("不支持此货币: %s", to)
	}
}

func transformCurrencyFromCHF(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyCHF {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	}
	switch to {
	case CurrencyUSD:
		return amount * 1.087, nil
	case CurrencyCNY:
		return amount * 7.609, nil
	case CurrencyEUR:
		return amount * 0.924, nil
	case CurrencyJPY:
		return amount * 119.57, nil
	case CurrencyGBP:
		return amount * 0.815, nil
	case CurrencyAUD:
		return amount * 1.468, nil
	case CurrencyCAD:
		return amount * 1.359, nil
	case CurrencyHKD:
		return amount * 8.476, nil
	case CurrencySGD:
		return amount * 1.468, nil
	default:
		return 0, fmt.Errorf("不支持此货币: %s", to)
	}
}

func transformCurrencyFromHKD(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencyHKD {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	}
	switch to {
	case CurrencyUSD:
		return amount * 0.128, nil
	case CurrencyCNY:
		return amount * 0.897, nil
	case CurrencyEUR:
		return amount * 0.109, nil
	case CurrencyJPY:
		return amount * 14.1, nil
	case CurrencyGBP:
		return amount * 0.096, nil
	case CurrencyAUD:
		return amount * 0.173, nil
	case CurrencyCAD:
		return amount * 0.16, nil
	case CurrencyCHF:
		return amount * 0.118, nil
	case CurrencySGD:
		return amount * 0.173, nil
	default:
		return 0, fmt.Errorf("不支持此货币: %s", to)
	}
}

func transformCurrencyFromSGD(amount float64, from Currency, to Currency) (float64, error) {
	if from != CurrencySGD {
		return 0, fmt.Errorf("不支持此货币: %s", from)
	}
	switch to {
	case CurrencyUSD:
		return amount * 0.741, nil
	case CurrencyCNY:
		return amount * 5.185, nil
	case CurrencyEUR:
		return amount * 0.63, nil
	case CurrencyJPY:
		return amount * 81.48, nil
	case CurrencyGBP:
		return amount * 0.556, nil
	case CurrencyAUD:
		return amount * 1.0, nil
	case CurrencyCAD:
		return amount * 0.926, nil
	case CurrencyCHF:
		return amount * 0.681, nil
	case CurrencyHKD:
		return amount * 5.778, nil
	default:
		return 0, fmt.Errorf("不支持此货币: %s", to)
	}
}
