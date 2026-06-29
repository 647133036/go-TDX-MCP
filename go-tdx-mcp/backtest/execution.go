package backtest

type ExecutionModel interface {
	Execute(price, volume, cash float64, isBuy bool) (execPrice float64, execVolume float64, execCost float64)
}

type ImmediateExecution struct{}

func (e *ImmediateExecution) Execute(price, volume, cash float64, isBuy bool) (float64, float64, float64) {
	if isBuy {
		buyable := int(cash / price)
		if buyable <= 0 {
			return price, 0, 0
		}
		vol := float64(buyable)
		return price, vol, vol * price
	}
	return price, volume, volume * price
}

type TWAPExecution struct {
	Periods int
}

func (e *TWAPExecution) Execute(price, volume, cash float64, isBuy bool) (float64, float64, float64) {
	if e.Periods <= 0 {
		e.Periods = 1
	}
	avgPrice := price * (1 + 0.0001*float64(e.Periods-1)/float64(e.Periods))
	if isBuy {
		buyable := int(cash / avgPrice)
		if buyable <= 0 {
			return avgPrice, 0, 0
		}
		vol := float64(buyable)
		return avgPrice, vol, vol * avgPrice
	}
	return avgPrice, volume, volume * avgPrice
}

type VWAPExecution struct {
	Periods int
}

func (e *VWAPExecution) Execute(price, volume, cash float64, isBuy bool) (float64, float64, float64) {
	if e.Periods <= 0 {
		e.Periods = 1
	}
	avgPrice := price * (1 + 0.00005*float64(e.Periods-1)/float64(e.Periods))
	if isBuy {
		buyable := int(cash / avgPrice)
		if buyable <= 0 {
			return avgPrice, 0, 0
		}
		vol := float64(buyable)
		return avgPrice, vol, vol * avgPrice
	}
	return avgPrice, volume, volume * avgPrice
}

type LimitExecution struct {
	LimitPct float64
}

func (e *LimitExecution) Execute(price, volume, cash float64, isBuy bool) (float64, float64, float64) {
	limitPrice := price
	if isBuy {
		limitPrice = price * (1 - e.LimitPct)
	} else {
		limitPrice = price * (1 + e.LimitPct)
	}
	if isBuy {
		buyable := int(cash / limitPrice)
		if buyable <= 0 {
			return limitPrice, 0, 0
		}
		vol := float64(buyable)
		return limitPrice, vol, vol * limitPrice
	}
	return limitPrice, volume, volume * limitPrice
}
