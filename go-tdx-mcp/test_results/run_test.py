import urllib.request
import json
import datetime

results = []
passed = 0
failed = 0

tests = [
    ('Health Check', 'GET', 'http://localhost:8000/api/v1/health', None),
    ('Server Info', 'GET', 'http://localhost:8000/api/v1/server-info', None),
    ('Quotes GET', 'GET', 'http://localhost:8000/api/v1/quotes?codes=SZ000001,SH600000', None),
    ('Bars', 'GET', 'http://localhost:8000/api/v1/bars?code=000001&market=sz&period=day&count=5', None),
    ('Financial Report', 'GET', 'http://localhost:8000/api/v1/financial/report?code=000001&type=lrb', None),
    ('Announcements', 'GET', 'http://localhost:8000/api/v1/announcements?code=000001&count=3', None),
    ('Macro Data', 'GET', 'http://localhost:8000/api/v1/macro-data?indicator=CPI&count=3', None),
    ('Market Overview', 'GET', 'http://localhost:8000/api/v1/market-overview', None),
    ('Indicator List', 'GET', 'http://localhost:8000/api/v1/indicator/list', None),
    ('Indicator Compute', 'GET', 'http://localhost:8000/api/v1/indicator/compute_all?code=000001&market=sz&indicators=MACD,KDJ', None),
    ('Indicator Compute lowercase', 'GET', 'http://localhost:8000/api/v1/indicator/compute_all?code=000001&market=sz&indicators=macd,kdj', None),
    ('Indicator Compute MA', 'GET', 'http://localhost:8000/api/v1/indicator/compute_all?code=000001&market=sz&indicators=MA', None),
    ('Chanlun Analyze', 'GET', 'http://localhost:8000/api/v1/chanlun/analyze?code=000001&market=sz&period=day&count=30', None),
    ('Backtest Run', 'GET', 'http://localhost:8000/api/v1/backtest/run?code=000001&market=sz&strategy=ma_cross&count=30', None),
    ('News Sentiment', 'GET', 'http://localhost:8000/api/v1/news-sentiment?code=000001&count=3', None),
    ('Board List', 'GET', 'http://localhost:8000/api/v1/board/list?board_type=HY&top_n=5', None),
    ('Board Members', 'GET', 'http://localhost:8000/api/v1/board/members?board_symbol=BK1717&count=5', None),
    ('Board Ranking', 'GET', 'http://localhost:8000/api/v1/board/ranking?board_type=HY&top_n=5', None),
    ('Capital Flow', 'GET', 'http://localhost:8000/api/v1/capital-flow?code=000001&market=sz', None),
    ('Auction', 'GET', 'http://localhost:8000/api/v1/auction?code=000001&market=sz', None),
    ('Unusual', 'GET', 'http://localhost:8000/api/v1/unusual?market=sz&count=5', None),
    ('Market Stat', 'GET', 'http://localhost:8000/api/v1/market-stat', None),
    ('Symbol Info', 'GET', 'http://localhost:8000/api/v1/symbol-info?code=000001&market=sz', None),
    ('Quote List', 'GET', 'http://localhost:8000/api/v1/quote-list?count=5&sort_type=change_pct', None),
    ('Security Count', 'GET', 'http://localhost:8000/api/v1/security-count?market=SZ', None),
    ('Belong Board', 'GET', 'http://localhost:8000/api/v1/belong-board?code=000001&market=sz', None),
    ('Block Data', 'GET', 'http://localhost:8000/api/v1/block?filename=gn.dat', None),
    ('Ex Markets', 'GET', 'http://localhost:8000/api/v1/ex/markets', None),
    ('Ex Quote HK', 'GET', 'http://localhost:8000/api/v1/ex/quote?ex_market=hk&code=00700', None),
    ('Ex Bars HK', 'GET', 'http://localhost:8000/api/v1/ex/bars?ex_market=hk&code=00700', None),
    ('Offline Home', 'GET', 'http://localhost:8000/api/v1/offline/home', None),
    ('Scraper Crypto', 'GET', 'http://localhost:8000/api/v1/scraper/crypto', None),
    ('Scraper Fund Holding', 'GET', 'http://localhost:8000/api/v1/scraper/fund-holding?fund_code=169103&period=2025Q1', None),
    ('Northbound Flow', 'GET', 'http://localhost:8000/api/v1/scraper/northbound-flow', None),
]

now = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
output_lines = []
output_lines.append('========================================')
output_lines.append('go-tdx-mcp Function Test Results')
output_lines.append(f'Date: {now}')
output_lines.append('Server: http://localhost:8000')
output_lines.append('========================================')
output_lines.append('')

for name, method, url, body in tests:
    try:
        req = urllib.request.Request(url)
        if body:
            req.data = body.encode()
            req.add_header('Content-Type', 'application/json')
        
        resp = urllib.request.urlopen(req, timeout=30)
        data = resp.read().decode()
        status = resp.status
        
        if status == 200 and data.strip():
            result = '[PASS]'
            passed += 1
        else:
            result = '[FAIL]'
            failed += 1
            output_lines.append(f'{result} ({name}) {method} {url}')
            output_lines.append(f'  HTTP Status: {status}')
            output_lines.append(f'  Response: {data[:200]}')
            output_lines.append('')
            continue
    except Exception as e:
        result = '[FAIL]'
        failed += 1
        output_lines.append(f'{result} ({name}) {method} {url}')
        output_lines.append(f'  Error: {str(e)[:200]}')
        output_lines.append('')
        continue
    
    output_lines.append(f'{result} ({name}) {method} {url}')
    try:
        j = json.loads(data)
        if isinstance(j, dict):
            resp_preview = json.dumps(j, ensure_ascii=False)[:200]
        elif isinstance(j, list):
            resp_preview = json.dumps(j, ensure_ascii=False)[:200]
        else:
            resp_preview = data[:200]
    except:
        resp_preview = data[:200]
    output_lines.append(f'  HTTP Status: {status}')
    output_lines.append(f'  Response: {resp_preview}')
    output_lines.append('')

output_lines.append('========================================')
output_lines.append('SUMMARY')
output_lines.append('========================================')
output_lines.append(f'Total endpoints tested: {passed + failed}')
output_lines.append(f'Passed: {passed}')
output_lines.append(f'Failed: {failed}')
output_lines.append('')

output = '\n'.join(output_lines)

with open('test_results/function_test.txt', 'w') as f:
    f.write(output)

print(output)
