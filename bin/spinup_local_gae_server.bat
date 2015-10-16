
@echo off
REM This file will start the google app engine for you and will also start and smtp server on pot 8025.
start C:\Windows\system32\cmd.exe /k b:\go_appengine\dev_appserver.py b:\GDrive\Appdir\sdcrm\app --log_level debug --smtp_host=localhost --smtp_port=8025 --clear_datastore --skip_sdk_update_check true

start python SMTPDebugginsServer.py