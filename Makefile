include ../../config/justchess.env
export

normal:
	go run -race cmd/justchess/main.go

dlv:
	dlv --headless --listen localhost:40000 debug cmd/justchess/main.go

gdlv:
	gdlv connect localhost:40000