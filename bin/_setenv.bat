
@echo off

REM Entry point where the whole dev environment is set
REM Place the following line into the cmd promt shortcut and point to this setenv.bat
REM D:\Windows\system32\cmd.exe /k E:\Dev\WorkSpace\setenv.bat

set XDATDOCSDIR=b:\GDrive\Appdir\sdcrm
set RELATIVEPATH=bin\alias.txt

doskey /MACROFILE="%XDATDOCSDIR%\%RELATIVEPATH%"
REM set GOPATH=b:\GDrive\Appdir\sdcrm\goscripts
set PATH=%XDATDOCSDIR%\bin; %PATH%

prompt [SDCRM]$P$_$_$G

pushd %XDATDOCSDIR%\app
color 3B
REM color code 80 might have to be changed if switched to new system.(BG=128/128/128, FG=0/43/54
