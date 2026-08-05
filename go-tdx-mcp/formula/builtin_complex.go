package formula

import (
	"fmt"

	"github.com/tdx/go-tdx-mcp/indicator"
)

// registerComplexBuiltin registers compound indicator functions that reuse the
// indicator engine. Multi-output indicators expose their primary line under the
// function name and companion lines under *_SUFFIX names (e.g. MACD_DEA,
// KDJ_D, BOLL_UPPER, DMI_ADX, TRIX_MA, BRAR_AR).
func (r *FunctionRegistry) registerComplexBuiltin() {
	// Single-output indicators
	r.Register("ATR", "indicator", "1~2", "平均真实波幅 ATR(N)", fnATR)
	r.Register("ROC", "indicator", "1~2", "变动率 ROC(N)", fnROC)
	r.Register("PSY", "indicator", "1~2", "心理线 PSY(N)", fnPSY)
	r.Register("EXPMA", "indicator", "1~2", "指数均线 EXPMA(N)", fnEXPMA)
	r.Register("MTM", "indicator", "1~2", "动量指标 MTM(N)", fnMTM)
	r.Register("DPO", "indicator", "1~2", "区间震荡线 DPO(N)", fnDPO)
	r.Register("OBV", "indicator", "0", "能量潮 OBV", fnOBV)
	r.Register("VWAP", "indicator", "0", "成交量加权均价 VWAP", fnVWAP)
	r.Register("SAR", "indicator", "0~2", "抛物线转向 SAR", fnSAR)
	r.Register("BBI", "indicator", "0~4", "多空指标 BBI", fnBBI)
	r.Register("ASI", "indicator", "1~2", "振动升降指标 ASI(N)", fnASI)
	r.Register("AROON", "indicator", "1~2", "阿隆指标 AROON(N)", fnAROON)
	r.Register("RSI", "indicator", "1~2", "相对强弱指标 RSI(N)", fnRSI)
	r.Register("CCI", "indicator", "1~2", "顺势指标 CCI(N)", fnCCI)
	r.Register("WR", "indicator", "1~2", "威廉指标 WR(N)", fnWR)
	r.Register("BIAS", "indicator", "1~2", "乖离率 BIAS(N)", fnBIAS)
	r.Register("VR", "indicator", "1~2", "成交量比率 VR(N)", fnVR)
	r.Register("MFI", "indicator", "1~2", "资金流量指标 MFI(N)", fnMFI)
	r.Register("EMV", "indicator", "1~2", "简易波动指标 EMV(N)", fnEMV)

	// Multi-output indicators: primary line + companion lines
	r.Register("MACD", "indicator", "0~3", "指数平滑异同平均（主输出 DIF）", fnMACD)
	r.Register("MACD_DEA", "indicator", "0~3", "MACD 信号线 DEA", fnMACDDEA)
	r.Register("MACD_MACD", "indicator", "0~3", "MACD 柱状图（2*(DIF-DEA)）", fnMACDMACD)

	r.Register("KDJ", "indicator", "0~3", "随机指标（主输出 K）", fnKDJ)
	r.Register("KDJ_D", "indicator", "0~3", "KDJ D 线", fnKDJD)
	r.Register("KDJ_J", "indicator", "0~3", "KDJ J 线", fnKDJJ)

	r.Register("BOLL", "indicator", "0~2", "布林带（主输出 MID）", fnBOLL)
	r.Register("BOLL_UPPER", "indicator", "0~2", "布林带上轨", fnBOLLUpper)
	r.Register("BOLL_LOWER", "indicator", "0~2", "布林带下轨", fnBOLLLower)

	r.Register("DMI", "indicator", "0~2", "动向指标（主输出 PDI）", fnDMI)
	r.Register("DMI_MDI", "indicator", "0~2", "DMI MDI 线", fnDMIMDI)
	r.Register("DMI_ADX", "indicator", "0~2", "DMI ADX 线", fnDMIADX)
	r.Register("DMI_ADXR", "indicator", "0~2", "DMI ADXR 线", fnDMIADXR)

	r.Register("TRIX", "indicator", "0~2", "三重指数平滑均线（主输出 TRIX）", fnTRIX)
	r.Register("TRIX_MA", "indicator", "0~2", "TRIX 移动平均线", fnTRIXMA)

	r.Register("BRAR", "indicator", "1~2", "BRAR 人气意愿指标（主输出 BR）", fnBRAR)
	r.Register("BRAR_AR", "indicator", "1~2", "BRAR 人气指标 AR", fnBRARAR)
}

// intArgOr returns the int arg at idx if present, otherwise def.
func intArgOr(args []*Value, idx, def int) (int, error) {
	if idx >= len(args) {
		return def, nil
	}
	a := args[idx]
	if a == nil || a.IsArray || a.IsString || a.IsDraw {
		return 0, NewEvalError(fmt.Sprintf("argument %d must be a number", idx+1))
	}
	return int(a.Single), nil
}

// floatArgOr returns the float arg at idx if present, otherwise def.
func floatArgOr(args []*Value, idx int, def float64) (float64, error) {
	if idx >= len(args) {
		return def, nil
	}
	a := args[idx]
	if a == nil || a.IsArray || a.IsString || a.IsDraw {
		return 0, NewEvalError(fmt.Sprintf("argument %d must be a number", idx+1))
	}
	return a.Single, nil
}

func posIntOr(args []*Value, idx, def int, name string) (int, error) {
	n, err := intArgOr(args, idx, def)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, NewEvalError(fmt.Sprintf("%s period must be positive", name))
	}
	return n, nil
}

func fnATR(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 14, "ATR")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.ATR(bars, n).Values), nil
}

func fnROC(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 12, "ROC")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.ROC(bars, n).Values), nil
}

func fnPSY(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 12, "PSY")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.PSY(bars, n).Values), nil
}

func fnEXPMA(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 12, "EXPMA")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.EXPMA(bars, n).Values), nil
}

func fnMTM(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 12, "MTM")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.MTM(bars, n).Values), nil
}

func fnDPO(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 20, "DPO")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.DPO(bars, n).Values), nil
}

func fnOBV(args []*Value, bars []indicator.Bar) (*Value, error) {
	if err := expectZeroArgs(args, "OBV"); err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.OBV(bars).Values), nil
}

func fnVWAP(args []*Value, bars []indicator.Bar) (*Value, error) {
	if err := expectZeroArgs(args, "VWAP"); err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.VWAP(bars).Values), nil
}

func fnSAR(args []*Value, bars []indicator.Bar) (*Value, error) {
	step, err := floatArgOr(args, 0, 0.02)
	if err != nil {
		return nil, err
	}
	afMax, err := floatArgOr(args, 1, 0.2)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.SAR(bars, step, afMax).Values), nil
}

func fnBBI(args []*Value, bars []indicator.Bar) (*Value, error) {
	n1, err := posIntOr(args, 0, 3, "BBI")
	if err != nil {
		return nil, err
	}
	n2, err := posIntOr(args, 1, 6, "BBI")
	if err != nil {
		return nil, err
	}
	n3, err := posIntOr(args, 2, 12, "BBI")
	if err != nil {
		return nil, err
	}
	n4, err := posIntOr(args, 3, 24, "BBI")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.BBI(bars, n1, n2, n3, n4).Values), nil
}

func fnASI(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 26, "ASI")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.ASI(bars, n).Values), nil
}

func fnAROON(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 25, "AROON")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.AROON(bars, n).Values), nil
}

func fnRSI(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 14, "RSI")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.RSI(bars, n).Values), nil
}

func fnCCI(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 14, "CCI")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.CCI(bars, n).Values), nil
}

func fnWR(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 14, "WR")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.WR(bars, n).Values), nil
}

func fnBIAS(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 6, "BIAS")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.BIAS(bars, n).Values), nil
}

func fnVR(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 26, "VR")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.VR(bars, n).Values), nil
}

func fnMFI(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 14, "MFI")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.MFI(bars, n).Values), nil
}

func fnEMV(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 14, "EMV")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.EMV(bars, n).Values), nil
}

func fnMACD(args []*Value, bars []indicator.Bar) (*Value, error) {
	fast, slow, signal, err := macdParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.MACD(bars, fast, slow, signal).Values), nil
}

func fnMACDDEA(args []*Value, bars []indicator.Bar) (*Value, error) {
	fast, slow, signal, err := macdParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.MACD(bars, fast, slow, signal).Line2), nil
}

func fnMACDMACD(args []*Value, bars []indicator.Bar) (*Value, error) {
	fast, slow, signal, err := macdParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.MACD(bars, fast, slow, signal).Line3), nil
}

func macdParams(args []*Value) (int, int, int, error) {
	fast, err := posIntOr(args, 0, 12, "MACD")
	if err != nil {
		return 0, 0, 0, err
	}
	slow, err := posIntOr(args, 1, 26, "MACD")
	if err != nil {
		return 0, 0, 0, err
	}
	signal, err := posIntOr(args, 2, 9, "MACD")
	if err != nil {
		return 0, 0, 0, err
	}
	return fast, slow, signal, nil
}

func fnKDJ(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m1, m2, err := kdjParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.KDJ(bars, n, m1, m2).Values), nil
}

func fnKDJD(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m1, m2, err := kdjParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.KDJ(bars, n, m1, m2).Line2), nil
}

func fnKDJJ(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m1, m2, err := kdjParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.KDJ(bars, n, m1, m2).Line3), nil
}

func kdjParams(args []*Value) (int, int, int, error) {
	n, err := posIntOr(args, 0, 9, "KDJ")
	if err != nil {
		return 0, 0, 0, err
	}
	m1, err := posIntOr(args, 1, 3, "KDJ")
	if err != nil {
		return 0, 0, 0, err
	}
	m2, err := posIntOr(args, 2, 3, "KDJ")
	if err != nil {
		return 0, 0, 0, err
	}
	return n, m1, m2, nil
}

func fnBOLL(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, p, err := bollParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.BOLL(bars, n, p).Values), nil
}

func fnBOLLUpper(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, p, err := bollParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.BOLL(bars, n, p).Line2), nil
}

func fnBOLLLower(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, p, err := bollParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.BOLL(bars, n, p).Line3), nil
}

func bollParams(args []*Value) (int, float64, error) {
	n, err := posIntOr(args, 0, 20, "BOLL")
	if err != nil {
		return 0, 0, err
	}
	p, err := floatArgOr(args, 1, 2.0)
	if err != nil {
		return 0, 0, err
	}
	return n, p, nil
}

func fnDMI(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m, err := dmiParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.DMI(bars, n, m).Values), nil
}

func fnDMIMDI(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m, err := dmiParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.DMI(bars, n, m).Line2), nil
}

func fnDMIADX(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m, err := dmiParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.DMI(bars, n, m).Line3), nil
}

func fnDMIADXR(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m, err := dmiParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.DMI(bars, n, m).Data["ADXR"]), nil
}

func dmiParams(args []*Value) (int, int, error) {
	n, err := posIntOr(args, 0, 14, "DMI")
	if err != nil {
		return 0, 0, err
	}
	m, err := posIntOr(args, 1, 6, "DMI")
	if err != nil {
		return 0, 0, err
	}
	return n, m, nil
}

func fnTRIX(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m, err := trixParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.TRIX(bars, n, m).Values), nil
}

func fnTRIXMA(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, m, err := trixParams(args)
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.TRIX(bars, n, m).Line2), nil
}

func trixParams(args []*Value) (int, int, error) {
	n, err := posIntOr(args, 0, 12, "TRIX")
	if err != nil {
		return 0, 0, err
	}
	m, err := posIntOr(args, 1, 9, "TRIX")
	if err != nil {
		return 0, 0, err
	}
	return n, m, nil
}

func fnBRAR(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 26, "BRAR")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.BRAR(bars, n).Values), nil
}

func fnBRARAR(args []*Value, bars []indicator.Bar) (*Value, error) {
	n, err := posIntOr(args, 0, 26, "BRAR")
	if err != nil {
		return nil, err
	}
	return NewArrayValue(indicator.BRAR(bars, n).Line2), nil
}

func expectZeroArgs(args []*Value, name string) error {
	if len(args) != 0 {
		return NewEvalError(fmt.Sprintf("%s requires 0 arguments", name))
	}
	return nil
}
