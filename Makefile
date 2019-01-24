version="0.10.6"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"

clean:
	-rm tpflow

build-go:
	go build -o tello-ad src/service.go

build-go-arm:
	cd ./src;GOOS=linux GOARCH=arm GOARM=6 go build -o tello-ad src/service.go;cd ../

build-go-amd:
	GOOS=linux GOARCH=amd64 go build -o tello-ad src/service.go


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf tello-ad_$(version).tar.gz tello-ad VERSION

package-deb-doc-1:
	@echo "Packaging application as debian package"
	chmod a+x package/debian1/DEBIAN/*
	cp tello-ad package/debian1/opt/thingsplex/tello-ad
	cp VERSION package/debian1/opt/thingsplex/tello-ad
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"

package-deb-doc-2:
	@echo "Packaging application as debian package"
	chmod a+x package/debian2/DEBIAN/*
	cp tello-ad package/debian2/usr/bin/tello-ad
	cp VERSION package/debian2/opt/thingsplex/tello-ad
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"


tar-arm: build-js build-go-arm package-deb-doc-1
	@echo "The application was packaged into tar archive "

deb-arm : clean configure-arm build-go-arm package-deb-doc-1
	mv package/debian.deb package/build/tello-ad_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc-1
	mv debian.deb tello-ad_$(version)_amd64.deb

run :
	go run src/service.go -c testdata/var/config.json


.phony : clean
