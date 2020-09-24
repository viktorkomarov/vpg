dev : clean all

all : postgres 
	./postgres

postgres : main.go conn.go
	@go build -o postgres *.go

.PHONY : clean
clean : 
	@rm postgres || true