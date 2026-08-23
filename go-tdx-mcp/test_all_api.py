#!/usr/bin/env python3
import urllib.request, urllib.error, json, sys, time

BASE = "http://localhost:8050"
passed = 0
failed = 0
errors = []

def get(path, name, expect_keys=None):
    global passed, failed
    try:
        req = urllib.request.Request(BASE + path, method="GET")
        resp = urllib.request.urlopen(req, timeout=30)
        body = resp.read().decode()
        d = json.loads(body)
        if "error" in d and d["error"]:
            errors.append(f"FAIL {name}: error={d['error']}")
            failed += 1
        else:
            if expect_keys:
                for k in expect_keys:
                    if k not in d:
                        errors.append(f"FAIL {name}: missing key '{k}'")
                        failed += 1
                        break
                else:
                    passed += 1
            else:
                passed += 1
    except Exception as e:
        errors.append(f"FAIL {name}: {e}")
        failed += 1

def post(path, body, name, expect_keys=None):
    global passed, failed
    try:
        data = json.dumps(body).encode()
        req = urllib.request.Request(BASE + path, data=data, method="POST",
                                    headers={"Content-Type": "application/json"})
        resp = urllib.request.urlopen(req, timeout=60)
        resp_body = resp.read().decode()
        d = json.loads(resp_body)
        if "error" in d and d["error"]:
            errors.append(f"FAIL {name}: error={d['error']}")
            failed += 1
        else:
            passed += 1
    except Exception as e:
        errors.append(f"FAIL {name}: {e}")
        failed += 1

def expect_error(path, name):
    global passed, failed
    try:
        req = urllib.request.Request(BASE + path, method="GET")
        urllib.request.urlopen(req, timeout=30)
        errors.append(f"FAIL {name}: expected error but got success")
        failed += 1
    except urllib.error.HTTPError as e:
        if e.code == 400:
            passed += 1
        else:
            errors.append(f"FAIL {name}: HTTP Error {e.code}")
            failed += 1
    except Exception as e:
        errors.append(f"FAIL {name}: {e}")
        failed += 1

def post_expect_error(path, body, name):
    global passed, failed
    try:
        data = json.dumps(body).encode()
        req = urllib.request.Request(BASE + path, data=data, method="POST",
                                    headers={"Content-Type": "application/json"})
        urllib.request.urlopen(req, timeout=30)
        errors.append(f"FAIL {name}: expected error but got success")
        failed += 1
    except urllib.error.HTTPError as e:
        if e.code == 400:
            passed += 1
        else:
            errors.append(f"FAIL {name}: HTTP Error {e.code}")
            failed += 1
    except Exception as e:
        errors.append(f"FAIL {name}: {e}")
        failed += 1

print("=" * 60)
print("全量功能测试")
print("=" * 60)

# --- 1. 健康检查 ---
print("\n[健康检查]")
get("/api/v1/health", "health", ["status"])

# --- 2. 核心接口 ---
print("\n[核心接口]")
get("/api/v1/bars?code=000001&market=0&period=1&count=5", "bars daily", ["data"])
get("/api/v1/bars?code=000001&market=0&period=60&count=5", "bars 60min", ["data"])
get("/api/v1/bars?code=000001&market=0&period=300&count=5", "bars weekly", ["data"])
get("/api/v1/quotes?code=000001&market=0", "quotes", ["data"])
get("/api/v1/quote-list?count=3", "quote-list", ["data"])
get("/api/v1/minute?code=000001&market=0", "minute", ["data"])
get("/api/v1/transaction?code=000001&market=0&count=5", "transaction", ["data"])
get("/api/v1/financial/report?code=000001&type=lrb", "financial/report", ["data"])
get("/api/v1/announcements?code=000001&count=2", "announcements", ["data"])
get("/api/v1/macro-data?indicator=LPR", "macro-data", ["data"])
get("/api/v1/market-overview", "market-overview", ["data"])
get("/api/v1/symbol-info?code=000001&market=0", "symbol-info", ["data"])
get("/api/v1/bars/index?code=000001&period=1&count=5", "bars/index", ["data"])
get("/api/v1/server-info", "server-info", ["data"])
get("/api/v1/server/hosts", "server/hosts", ["hosts"])
get("/api/v1/finance?code=000001&market=0", "finance", ["data"])
get("/api/v1/security-count?market=SZ", "security-count", ["count"])
get("/api/v1/security/list?market=SZ&count=3&start=0", "security/list", ["data"])

# --- 3. 分析工具 ---
print("\n[分析工具]")
get("/api/v1/indicator/list", "indicator/list", ["indicators"])
get("/api/v1/indicator/compute?code=000001&market=0&period=1&count=30&indicator=MACD",
    "indicator/compute MACD", ["MACD"])
get("/api/v1/indicator/compute_all?code=000001&market=0&period=1&count=30",
    "indicator/compute_all", ["MACD", "KDJ", "RSI"])
get("/api/v1/chanlun/analyze?code=000001&market=0&period=d&count=500",
    "chanlun/analyze", ["BiList"])
get("/api/v1/backtest/run?code=000001&market=0&strategy=MA_CROSS&period=1&count=200",
    "backtest/run", ["total_return", "sharpe"])
get("/api/v1/backtest/run-all?code=000001&market=0&count=500",
    "backtest/run-all", ["results", "strategy_count"])
get("/api/v1/news-sentiment?code=000001&count=2",
    "news-sentiment", ["data"])

# POST analysis
print("\n[分析工具-POST]")
post("/api/v1/chanlun/multi",
     {"code": "000001", "market": 0, "config": {"bi_type": "old"},
      "levels": [{"name": "daily", "period": "1", "count": 500}]},
     "chanlun/multi", ["levels", "config"])
post("/api/v1/backtest/signal-scan",
     {"strategy": "ma_cross", "codes": ["000001", "600000"]},
     "backtest/signal-scan", ["rows"])
post("/api/v1/backtest/signal-rank",
     {"codes": ["000001", "600000"], "strategy": "ma_cross", "period": "1", "count": 200},
     "backtest/signal-rank", ["results"])

# --- 4. 板块与资金 ---
print("\n[板块与资金]")
get("/api/v1/board/list?board_type=HY&count=3", "board/list", ["data"])
get("/api/v1/board/members?board_symbol=BK1033&count=3", "board/members", ["data"])
get("/api/v1/board/ranking?board_type=HY&top_n=3", "board/ranking", ["data"])
get("/api/v1/board/change-ranking?board_type=HY", "board/change-ranking", ["data"])
get("/api/v1/board/summary?board_symbol=BK1033", "board/summary", ["data"])
get("/api/v1/capital-flow?code=000001&market=0", "capital-flow", ["data"])
get("/api/v1/auction?code=000001&market=0", "auction", ["data"])
get("/api/v1/unusual?market=0&count=3", "unusual", ["data"])
get("/api/v1/market-stat", "market-stat", ["data"])

# --- 5. 证券信息 ---
print("\n[证券信息]")
get("/api/v1/belong-board?code=000001&market=0", "belong-board", ["data"])
get("/api/v1/block?filename=block_gy.dat", "block", ["data"])

# --- 6. 公司/除权 ---
print("\n[公司/除权]")
get("/api/v1/company/category?code=000001&market=0", "company/category", ["data"])
get("/api/v1/xdxr?code=000001&market=0", "xdxr", ["data"])
get("/api/v1/volume-profile?code=000001&market=0", "volume-profile", ["profile"])
get("/api/v1/index/info?code=000001&market=1", "index/info", ["orders"])
get("/api/v1/index/momentum?code=000001&market=1", "index/momentum", ["values"])
get("/api/v1/history-orders?code=000001&market=0&date=20260821", "history-orders", ["orders"])
get("/api/v1/fund-flow?code=000001&market=0", "fund-flow", ["today_main_net_in"])
get("/api/v1/ex/quotes-list?category=31&count=3", "ex/quotes-list", ["data"])

# --- 7. 扩展市场 ---
print("\n[扩展市场]")
get("/api/v1/ex/markets", "ex/markets", ["markets"])
get("/api/v1/ex/bars?ex_market=HK_MAIN_BOARD&code=00700", "ex/bars", ["data"])
get("/api/v1/ex/quote?ex_market=HK_MAIN_BOARD&code=00700", "ex/quote", ["data"])
get("/api/v1/ex/quotes?ex_market=HK_MAIN_BOARD&codes=00700", "ex/quotes", ["data"])
get("/api/v1/ex/minute?ex_market=HK_MAIN_BOARD&code=00700", "ex/minute", ["data"])
get("/api/v1/ex/transaction?ex_market=HK_MAIN_BOARD&code=00700", "ex/transaction", ["data"])
get("/api/v1/ex/list?ex_market=HK_MAIN_BOARD&category=31&count=3", "ex/list", ["data"])
get("/api/v1/ex/chart-sampling?ex_market=HK_MAIN_BOARD&code=00700", "ex/chart-sampling", ["data"])
get("/api/v1/ex/transaction-all?ex_market=HK_MAIN_BOARD&code=00700", "ex/transaction-all", ["data"])

# --- 8. 离线 ---
print("\n[离线数据]")
get("/api/v1/offline/home", "offline/home", ["data"])
get("/api/v1/offline/daily?code=000001&year=2025&month=1", "offline/daily", ["data"])
get("/api/v1/offline/gbbq?code=000001&year=2025", "offline/gbbq", ["data"])

# --- 9. Scraper ---
print("\n[Scraper模块]")
get("/api/v1/scraper/sector-boards?board_type=HY&count=2", "scraper/sector-boards", ["data"])
get("/api/v1/scraper/margin-trade", "scraper/margin-trade", ["data"])

# --- 10. POST 回测 ---
print("\n[POST回测]")
post("/api/v1/backtest/run",
     {"strategy": "ma_cross", "code": "000001", "market": 0, "period": "1", "count": 200},
     "backtest/run POST", ["total_return"])
post("/api/v1/backtest/optimize",
     {"strategy": "ma_cross", "params": {"fast": [5, 10], "slow": [20, 30]},
      "code": "000001", "market": 0, "period": "1", "count": 200},
     "backtest/optimize", ["results"])
post("/api/v1/backtest/multi-strategy",
     {"items": [{"strategy": "ma_cross", "code": "000001", "market": 0}],
      "period": "1", "count": 200, "cash": 1000000},
     "backtest/multi-strategy", ["results"])
post("/api/v1/backtest/portfolio",
     {"strategy": "ma_cross", "period": "1", "count": 200, "cash": 1000000,
      "stocks": [{"code": "000001", "market": 0}]},
     "backtest/portfolio", ["results"])

# --- 11. 因子 ---
print("\n[因子分析]")
get("/api/v1/factor/list", "factor/list", ["factors"])
post("/api/v1/factor/compute",
     {"codes": ["000001", "600000"], "market": 0, "period": "1", "count": 500},
     "factor/compute", ["data"])

# --- 12. 错误处理 ---
print("\n[错误处理]")
expect_error("/api/v1/bars?code=000001", "bars no-market")
post_expect_error("/api/v1/chanlun/multi", {}, "chanlun/multi empty")
expect_error("/api/v1/xdxr", "xdxr no-code")

# --- 13. POST 缠论多级别 ---
print("\n[POST 缠论多级别联立]")
post("/api/v1/chanlun/multi",
     {"code": "000001", "market": 0,
      "config": {"bi_type": "old", "zs_type": "strict", "fx_strict": True},
      "levels": [
          {"name": "daily", "period": "1", "count": 500},
          {"name": "week", "period": "5", "count": 200}
      ]},
     "chanlun/multi 2-levels", ["levels"])

# --- Summary ---
print("\n" + "=" * 60)
print(f"PASS: {passed}")
print(f"FAIL: {failed}")
print(f"TOTAL: {passed + failed}")
print("=" * 60)
if errors:
    print("\n失败详情:")
    for e in errors:
        print(f"  {e}")

sys.exit(0 if failed == 0 else 1)