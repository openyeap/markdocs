@echo off
setlocal

FOR /F "delims=" %%i IN ("%cd%") DO (
    set name=%%~ni
) 
cd target
mv  *.html assets docs html OEBPS

zip  %name%.epub  OEBPS/*  META-INF/* mimetype

cd ..
