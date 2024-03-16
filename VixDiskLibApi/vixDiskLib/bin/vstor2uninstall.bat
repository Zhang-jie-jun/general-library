@echo off
rem This batch file uninstalls the vstor2-mntapi20-shared service.
rem
rem Check the system type
rem
rem Check if the service is installed
rem
set VSTOR=vstor2-mntapi20-shared
sc query %VSTOR% > nul
if errorlevel 1 (
    echo "Service not found. Quitting"
    exit /b 1
)
rem
rem The service is installed, remove it.
rem
echo "Service exists, removing it."
rundll32.exe setupapi.dll,InstallHinfSection DefaultUninstall 132 %~dp0\AMD64\vstor2.inf
rem The service should be cleaned up.
sc query %VSTOR% > nul
if errorlevel 1 (
    echo "Uninstall successfully."
    exit /b 0
)
echo "Failed to uninstall."
exit /b 2
