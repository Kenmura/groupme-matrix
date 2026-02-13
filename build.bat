@echo off
go build -tags nocrypto .
if %errorlevel% neq 0 exit /b %errorlevel%
echo Build successful! Created groupme.exe
pause
