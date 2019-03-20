.PHONY: assets

assets:
	cd internal && go-bindata -pkg internal config.ini