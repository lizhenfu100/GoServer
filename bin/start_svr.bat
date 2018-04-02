start ./mongodb/mongod.exe --dbpath ./db --auth
start ./zookeeper.exe
rem start ./svr_battle.exe 1
rem start ./svr_battle.exe 2
start ./svr_center.exe
start ./svr_cross.exe
start ./svr_gateway.exe
start ./svr_friend.exe
rem start ./svr_login.exe
start ./svr_game.exe
rem start ./svr_client.exe