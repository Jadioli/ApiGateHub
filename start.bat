@echo off
setlocal

if not exist ".env" (
  copy /Y ".env.example" ".env" >nul
)

if not exist "web\dist\index.html" (
  echo Building frontend...
  pushd web
  call npm ci
  if errorlevel 1 (
    popd
    exit /b 1
  )
  call npm run build
  if errorlevel 1 (
    popd
    exit /b 1
  )
  popd
)

echo Building backend...
go build -o apihub.exe ./cmd/server
if errorlevel 1 exit /b 1

echo ApiHub listening on http://localhost:9011
apihub.exe
