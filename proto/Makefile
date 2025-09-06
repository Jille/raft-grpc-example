service.pb.go: service.proto
	protoc service.proto --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

force:
	rm -f service.pb.go
	make service.pb.go
