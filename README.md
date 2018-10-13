## dev
```bash
# 1. edit proto
$ vim edit greet/greetpb/greetproto

# 2. generate code
# 2-1. golang
$ protoc greet/greetpb/greet.proto --go_out=plugins=grpc:.
# 2-2. ruby
$ bundle exec grpc_tools_ruby_protoc --ruby_out=. --grpc_out=. greet/greetpb/greet.proto

# 3. implement server
# 3-1. golang
$ vim server.go
# 3-2. ruby
$ vim server.rb

# 4. implement server
# 4-1. golang
$ vim client.go
# 4-2. ruby
# $ vim client.rb ... まだ書いてない
```