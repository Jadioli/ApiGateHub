@echo off
setlocal

if not exist ".env" (
  copy /Y ".env.example" ".env" >nul
)

rem Frontend build caching notes:
rem - Previously we only built when web\dist\index.html was missing.
rem - That can go stale when web\src changes (dist still exists).
rem - We rebuild when any frontend input is newer than dist\index.html.

set "DIST_HTML=web\dist\index.html"
set "NEED_BUILD_FRONTEND=0"

if not exist "%DIST_HTML%" (
  set "NEED_BUILD_FRONTEND=1"
) else (
  powershell -NoProfile -ExecutionPolicy Bypass -Command ^
    "$dist = Get-Item '%DIST_HTML%';" ^
    "$paths = @('web/src','web/public','web/scripts') | Where-Object { Test-Path $_ };" ^
    "$files = Get-ChildItem -Recurse -File -ErrorAction SilentlyContinue $paths;" ^
    "$extra = @('web/package.json','web/package-lock.json','web/npm-shrinkwrap.json') | Where-Object { Test-Path $_ } | ForEach-Object { Get-Item $_ };" ^
    "$newer = $files + $extra | Where-Object { $_.LastWriteTime -gt $dist.LastWriteTime } | Select-Object -First 1;" ^
    "if ($newer) { exit 0 } else { exit 1 }"
  if %errorlevel%==0 set "NEED_BUILD_FRONTEND=1"
)

if "%NEED_BUILD_FRONTEND%"=="1" (
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
