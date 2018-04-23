build:
	go build -o ./wort ./cmd/wort
	svelte compile --format iife ./svelte/MainApp.html > ./html/static/js/MainApp.js
