import akshare as ak 

df = ak.stock_info_a_code_name()

df.to_csv("ag3.csv",index=False,encoding="utf_8_sig")

print("csv already save!")