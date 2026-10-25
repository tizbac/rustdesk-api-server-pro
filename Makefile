main_output=build
frontend=soybean-admin
frontend_dist=${frontend}/dist

build: clean
	go build -C backend -o ../${main_output}/
	cd ${frontend} && pnpm build && cp -R dist ${main_output}/

# builds the rustdesk web client, it needs flutter and yarn (see rustdesk/build_web.sh)
web:
	./rustdesk/build_web.sh
	mkdir -p ${main_output}
	rm -rf ${main_output}/web
	cp -R rustdesk/flutter/build/web ${main_output}/web

clean:
	rm -rf ${main_output}
	rm -rf ${frontend_dist}
	go clean
