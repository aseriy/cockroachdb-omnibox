.PHONY: build ui clean

build: ui
	go build

ui:
	cd ui && npm install && npm run build

clean:
	rm -rf ui/dist ui/node_modules omnibox-visualize
