## 读取股票数据
使用 func ReadDataTable(filename string) (table *DataTable) 读取股票数据。
然后做统计分析。分析每一个股票的各年度中位数价格，1/5低价，1/5高价。

### 数据格式
date        	open    high    close   low     volumn          deal money      pow
2004-01-02	16.006	16.406	16.160	15.883	11565225.000	121756888.000	1.539

date 日期
open 开盘价
high 最高价
close 收盘价
low 最低价
volume成交量（手）
deal money 成交额
pow 权

