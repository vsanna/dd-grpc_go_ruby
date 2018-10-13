#!/bin/bash

protoc greet/greetpb/greet.proto --go_out=plugins=grpc:.
bundle exec grpc_tools_ruby_protoc --ruby_out=. --grpc_out=. greet/greetpb/greet.proto