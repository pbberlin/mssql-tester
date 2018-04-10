# mssql-tester

mssql-tester executes a configurable query against a MS SQL Server.

The query is repeated at configurable intervals.

The query result is logged to a file.

The connection can be encrypted, using a configurable *.pem file.

Use `example.json` to create your own `connect.json`  
containing all parameters.

`mssql-tester.exe`  starts the program in visible command window.

`runinvisble.bat`  starts the program invisibly (Windows only).

`mssql-tester.log` contains errors and the query results
