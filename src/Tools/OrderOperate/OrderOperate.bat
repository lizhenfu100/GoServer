
rem 要查询的订单号
SET GetOrder="0091804045675481 0081804105915898 0091804125946789"

rem 要修改的订单号
SET SetOrder=""

start ./OrderOperate.exe -ip "120.78.152.152" -port 7002 -g %GetOrder% -s %SetOrder%
