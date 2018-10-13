package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vsanna/grpc/greet/greetpb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello, I'm a client")

	// setup client
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("cannot connect: %v", err)
	}
	defer conn.Close()

	// grpcに変化させてリクエスト飛ばす
	c := greetpb.NewGreetServiceClient(conn)
	// fmt.Printf("created client: %f", c)
	/*
		Hello, I'm a client
		created client: &{%!f(*grpc.ClientConn=&{0xc420072280 0x1130970 localshot:50051 {passthrough  localshot:50051} localshot:50051 {<nil> <nil> <nil> <nil> {120000000000} false true 0 <nil>  {grpc-go/1.16.0-dev 0x133e780 false [] <nil> <nil> {0 0 false} <nil> 0 0 32768 32768 0 <nil>} [] <nil> 0x1686a00 false 0 false true} 0xc4201204a0 {<nil> <nil> 0x133e780 0} 0xc4200722c0 0xc420072240 {{0 0} 0 0 0 0} {<nil> map[] <nil>}  map[] {0 0 false}   [] <nil> {<nil>} 0 0xc4200284e0})}%
	*/

	// doUnary(c)
	// doServerStreaming(c)
	doClientStreaming(c)
	// doBiDiStreaming(c)
	// doErrorUnary(c)

	// doUnaryWithDeadline(c, 5*time.Second)
	// doUnaryWithDeadline(c, 1*time.Second)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("starting to do a Unary RPC...")

	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "taro",
			LastName:  "yamada",
		},
	}
	res, err := c.Greet(context.Background(), req)

	if err != nil {
		log.Fatalf("error white calling greet RPC: %v", err)
	}
	log.Printf("response from Greet: %v", res)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("starting to do a ServerStreaming RPC...")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "ryu",
			LastName:  "ishikawa",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	resStream, err := c.GreetManyTimes(ctx, req)

	if err != nil {
		log.Fatalf("response from GreetManyTimes: %v", err)
	}

	for {
		// ここでちゃんと待ってるな. fmt.Println("高速?")としてもfor分しか走らない
		msg, err := resStream.Recv()
		if err == io.EOF {
			// finish stream
			break
		}
		if err != nil {
			log.Fatalf("error while readinng stream %v", err)
		}

		log.Printf("response from GreetManyTimes: %v", msg)
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("starting to do a ClientStreaming RPC...")

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "name1",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "name2",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "name3",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}

	for _, req := range requests {
		stream.Send(req)
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving LongGreetResponse: %v", err)
	}

	fmt.Printf("LongGreet Response: %v", res)
}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	// we create a stream by invoking the client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	requests := []*greetpb.GreetEveryoneRequest{}
	for i := 0; i < 10; i++ {
		requests = append(requests, &greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: fmt.Sprintf("name%d", i),
			},
		})
	}

	// we send a bunch of messages to the client (go routine)
	waitc := make(chan struct{})
	go func() {
		for _, req := range requests {
			stream.Send(req)
			time.Sleep(300 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// we receive a bunch of messages from the client(go routine)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				close(waitc)
			}

			fmt.Printf("Received: %v\n", res.GetResult())
		}
	}()

	// block untile everythinng is done
	<-waitc
}

func doErrorUnary(c greetpb.GreetServiceClient) {
	// correct call
	doErrorCall(c, 10)

	// error call
	doErrorCall(c, -10)
}

func doErrorCall(c greetpb.GreetServiceClient, number int32) {
	res, err := c.SquareRoot(context.Background(), &greetpb.SquareRootRequest{Number: number})
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC(user error)
			fmt.Println("Message: ", respErr.Message())
			fmt.Println("Code:    ", respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number!\n")
			}
		} else {
			// 通信そのもののエラー
			log.Fatalf("Big error calling SquareRoot: %v", err)
		}
	}

	fmt.Printf("Result of square root of %v: %v\n", number, res.GetNumberRoot())
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	req := &greetpb.GreetDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "ryu",
			LastName:  "ishikawa",
		},
	}

	// context.Background()をそのまま渡すのではなく、WithTimeoutにして渡す
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetDeadline(ctx, req)

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was Exceeded")
			} else {
				fmt.Println("unexpected error: %v", err)
			}
		} else {
			log.Fatalf("error while calling GreetingDeadline RPC: %v", err)
		}
	}

	log.Printf("Response from GreetDeadline: %v", res)
}
