package scraper

import (
	"strings"
	"testing"
)

const testGovernmentHTML = `
<html>
<body>
<tr><td class="sideColumn">&nbsp;</td><td class="left bold first noWrap"><a href="/indices/shanghai-composite" title="上证指数">上证指数</a></td><td class="lastNum pid-40820-last">3,900.35</td><td class="chg greenFont pid-40820-pc">+21.92</td><td class="chgPer greenFont pid-40820-pcp">+0.57%</td></tr>
<tr><td class="sideColumn">&nbsp;</td><td class="left bold first noWrap"><a href="/indices/ftse-china-a50" title="富时中国A50指数">富时中国A50指数</a></td><td class="lastNum pid-28930-last">14,964.53</td><td class="chg redFont pid-28930-pc">-11.50</td><td class="chgPer redFont pid-28930-pcp">-0.08%</td></tr>
<tr><td class="sideColumn">&nbsp;</td><td class="left bold first noWrap"><a href="/indices/hang-sen-40" title="香港恒生指数">香港恒生指数</a></td><td class="lastNum pid-179-last">25,530.28</td><td class="chg redFont pid-179-pc">-385.54</td><td class="chgPer redFont pid-179-pcp">-1.49%</td></tr>
</body>
</html>
`

func TestParseGovernmentHTML(t *testing.T) {
	quotes := parseGovernmentHTML(testGovernmentHTML)

	if len(quotes) != 3 {
		t.Fatalf("expected 3 quotes, got %d", len(quotes))
	}

	tests := []struct {
		name      string
		last      float64
		change    float64
		changePct float64
	}{
		{"上证指数", 3900.35, 21.92, 0.57},
		{"富时中国A50指数", 14964.53, -11.50, -0.08},
		{"香港恒生指数", 25530.28, -385.54, -1.49},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, q := range quotes {
				if strings.Contains(q.Name, tt.name) {
					if q.Last != tt.last {
						t.Errorf("%s Last = %v, want %v", tt.name, q.Last, tt.last)
					}
					if q.Change != tt.change {
						t.Errorf("%s Change = %v, want %v", tt.name, q.Change, tt.change)
					}
					if q.ChangePct != tt.changePct {
						t.Errorf("%s ChangePct = %v, want %v", tt.name, q.ChangePct, tt.changePct)
					}
				}
			}
		})
	}
}
