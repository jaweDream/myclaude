@echo off
setlocal enabledelayedexpansion

set "EXIT_CODE=0"
set "REPO=cexll/myclaude"
set "VERSION=latest"
set "OS=windows"

call :detect_arch
if errorlevel 1 goto :fail

set "BINARY_NAME=codex-wrapper-%OS%-%ARCH%.exe"
set "URL=https://github.com/%REPO%/releases/%VERSION%/download/%BINARY_NAME%"
set "TEMP_FILE=%TEMP%\codex-wrapper-%ARCH%-%RANDOM%.exe"
set "DEST_DIR=%USERPROFILE%\bin"
set "DEST=%DEST_DIR%\codex-wrapper.exe"

echo Downloading codex-wrapper for %ARCH% ...
echo   %URL%
call :download
if errorlevel 1 goto :fail

if not exist "%TEMP_FILE%" (
    echo ERROR: download failed to produce "%TEMP_FILE%".
    goto :fail
)

echo Installing to "%DEST%" ...
if not exist "%DEST_DIR%" (
    mkdir "%DEST_DIR%" >nul 2>nul || goto :fail
)

move /y "%TEMP_FILE%" "%DEST%" >nul 2>nul
if errorlevel 1 (
    echo ERROR: unable to place file in "%DEST%".
    goto :fail
)

"%DEST%" --version >nul 2>nul
if errorlevel 1 (
    echo ERROR: installation verification failed.
    goto :fail
)

echo.
echo codex-wrapper installed successfully at:
echo   %DEST%

set "PATH_CHECK=;%PATH%;"
echo !PATH_CHECK! | findstr /I /C:";%DEST_DIR%;" >nul
if errorlevel 1 (
    echo.
    echo %DEST_DIR% is not in your PATH.
    echo Add it for the current user with:
    echo   setx PATH "%%USERPROFILE%%\bin;%%PATH%%"
    echo Then restart your terminal to use codex-wrapper globally.
)

goto :cleanup

:detect_arch
set "ARCH=%PROCESSOR_ARCHITECTURE%"
if defined PROCESSOR_ARCHITEW6432 set "ARCH=%PROCESSOR_ARCHITEW6432%"

if /I "%ARCH%"=="AMD64" (
    set "ARCH=amd64"
    exit /b 0
) else if /I "%ARCH%"=="ARM64" (
    set "ARCH=arm64"
    exit /b 0
) else (
    echo ERROR: unsupported architecture "%ARCH%". 64-bit Windows on AMD64 or ARM64 is required.
    set "EXIT_CODE=1"
    exit /b 1
)

:download
where curl >nul 2>nul
if %errorlevel%==0 (
    echo Using curl ...
    curl -fL --retry 3 --connect-timeout 10 "%URL%" -o "%TEMP_FILE%"
    if errorlevel 1 (
        echo ERROR: curl download failed.
        set "EXIT_CODE=1"
        exit /b 1
    )
    exit /b 0
)

where powershell >nul 2>nul
if %errorlevel%==0 (
    echo Using PowerShell ...
    powershell -NoLogo -NoProfile -Command " $ErrorActionPreference='Stop'; try { [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor 3072 -bor 768 -bor 192 } catch {} ; $wc = New-Object System.Net.WebClient; $wc.DownloadFile('%URL%','%TEMP_FILE%') "
    if errorlevel 1 (
        echo ERROR: PowerShell download failed.
        set "EXIT_CODE=1"
        exit /b 1
    )
    exit /b 0
)

echo ERROR: neither curl nor PowerShell is available to download the installer.
set "EXIT_CODE=1"
exit /b 1

:fail
echo Installation failed.
set "EXIT_CODE=1"
goto :cleanup

:cleanup
if exist "%TEMP_FILE%" del /f /q "%TEMP_FILE%" >nul 2>nul
set "CODE=%EXIT_CODE%"
endlocal & exit /b %CODE%
