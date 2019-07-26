start ./mongodb/mongod.exe --port 27017 --dbpath ./db --auth
rem start ./mongodb/mongod.exe --port 27018 --dbpath ./db2 --auth
rem start ./mongodb/mongod.exe --master --oplogSize 4096 --dbpath ./db --auth
rem start ./mongodb/mongod.exe --slave --source 192.168.1.111:27017 --dbpath ./db --auth -port 27018

rem start ./mongodb/mongo.exe --port 27018