mkdir bin
rsrc -ico quail-view.ico -manifest quail-view.exe.manifest
copy /y quail-view.exe.manifest bin\quail-view.exe.manifest
go build -buildmode=pie -ldflags="-s -w" -o quail-view.exe main.go
move quail-view.exe bin/quail-view.exe
cd bin && quail-view.exe c:\games\eq\rebuildeq\luc.eqg