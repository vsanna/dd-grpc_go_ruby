require 'bundler'
Bundler.require

require_relative './greet/greetpb/greet_services_pb.rb'

# serviceの実装を作る
# memo: requestのmock: req = Greet::GreetRequest.new(greeting: Greet::Greeting.new(first_name: 'hofe', last_name: "geho"))
include GRPC::Core::StatusCodes
class GreetService < Greet::GreetService::Service
	def greet(greet_request, _unused_call)
		result = "Hello! #{greet_request.greeting.first_name} #{greet_request.greeting.last_name}"
		Greet::GreetResponse.new(result: result)
	end

	def square_root(square_root, _unused_call)
		number = square_root.number
		Greet::SquareRootResponse.new(number_root: Math.sqrt(number))
	rescue => e
		puts e
		raise GRPC::BadStatus.new_status_exception(INVALID_ARGUMENT)
	end

	# server streamingではenumeratorを返す
	def greet_many_times(request, _unused_call)
		# return [Greet::GreetManyTimesResponse.new(result: "hoge")] unless block_given?
		result = "Hello! #{request.greeting.first_name} #{request.greeting.last_name}"	

		10.times.map do |i|
			Greet::GreetManyTimesResponse.new(result: "#{i}: #{result}")
		end
	end

	# client streamingはstreamを受け取ってそこから1個のresponseを作って返す
	# 第2引数なし
	def long_greet(stream)
		result = ''
		stream.each_remote_read do |req|
			puts "get req: #{req}"
			result += "Hello! #{req.greeting.first_name} #{req.greeting.last_name}" 
		end
		Greet::LongGreetResponse.new(result: result)
	end
end

class GreetServer
	class << self
		def start
		  start_grpc_server
		end
	
		private

		def start_grpc_server
		  @server = GRPC::RpcServer.new
		  @server.add_http2_port("0.0.0.0:50052", :this_port_is_insecure)
		  @server.handle(GreetService)
		  @server.run_till_terminated
		end
	end
end


GreetServer.start