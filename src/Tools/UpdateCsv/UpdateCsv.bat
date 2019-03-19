rem 海外配置
set Files="Abroad/conf_svr.csv>csv/conf_svr.csv csv/game/const.csv csv/email/email.csv"
set Addrs=^
"52.14.1.205:7090 3.17.23.172:7090 18.221.148.84:7090 18.223.109.103:7090 18.216.113.27:7090 ^
13.229.215.168:7090 54.169.60.150:7090 13.229.102.23:7090 ^
18.185.80.202:7090 3.120.224.107:7090 ^
54.94.211.178:7090 18.231.107.243:7090"
start ./UpdateCsv.exe -file %Files% -addr %Addrs%

rem 国区配置
set Files="China/conf_svr.csv>csv/conf_svr.csv csv/game/const.csv csv/email/email.csv"
set Addrs=^
"39.96.196.250:7090 39.96.196.69:7090 ^
47.106.35.74:7090 47.107.41.54:7090 47.106.113.86:7090"
start ./UpdateCsv.exe -file %Files% -addr %Addrs%