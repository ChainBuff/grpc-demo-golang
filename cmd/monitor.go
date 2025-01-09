package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mr-tron/base58"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	pingInterval = 5 * time.Second
	pumpAddr     = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
)

// 统计ping的平均延迟
var (
	lastPingTime time.Time
	totalLatency time.Duration
	pingCount    int64
)

// 预编译字符串匹配
var initMintPattern = []byte("InitializeMint2")

var grpcClient pb.GeyserClient

// 监控代币事件
func monitorTokens(ctx context.Context, stream pb.Geyser_SubscribeClient) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		resp, err := stream.Recv()
		if err != nil {
			log.Printf("Stream error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// 接收响应时计算延迟
		if pong := resp.GetPong(); pong != nil {
			latency := time.Since(lastPingTime)
			totalLatency += latency
			pingCount++
			avgLatency := totalLatency / time.Duration(pingCount)

			log.Printf("Ping Count: %d, Latency: %v, Avg Latency: %v",
				pingCount, latency, avgLatency)
			continue
		}

		if data := resp.GetTransaction(); data != nil {
			meta := data.GetTransaction().GetMeta()

			logMessages := meta.GetLogMessages()
			if len(logMessages) == 0 {
				continue
			}

			// 检查是否是新代币创建
			// 使用 bytes.Contains 性能更好
			hasInitMint := false
			for _, log := range logMessages {
				if bytes.Contains([]byte(log), initMintPattern) {
					hasInitMint = true
					break
				}
			}

			if hasInitMint {
				accountKeys := data.GetTransaction().GetTransaction().GetMessage().GetAccountKeys()

				mintAddress := base58.Encode(accountKeys[1])
				bondingCurveAddress := base58.Encode(accountKeys[2])
				associatedBondingCurveAddress := base58.Encode(accountKeys[3])

				// 异步发送数据
				// TODO

				// 不必要的信息放到后边
				// 获取交易签名
				sig := base58.Encode(data.GetTransaction().GetTransaction().GetSignatures()[0])
				slot := data.GetSlot()

				// 异步打印
				go log.Printf("----- New Token Created -----\n"+
					"Signature: %s\n"+
					"Slot: %d\n"+
					"Mint: %s\n"+
					"Bonding Curve: %s\n"+
					"Associated Bonding Curve: %s\n\n",
					sig,
					slot,
					mintAddress,
					bondingCurveAddress,
					associatedBondingCurveAddress,
				)
			}
		}
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)

	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	grpcURL := os.Getenv("GRPC_URL")
	if grpcURL == "" {
		log.Fatalf("GRPC_URL is not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v, shutting down...", sig)
		cancel()
	}()

	// 设置 gRPC 连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
	}

	// 连接到 gRPC 服务
	conn, err := grpc.NewClient(grpcURL, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// 创建 Geyser 客户端
	grpcClient = pb.NewGeyserClient(conn)

	// 创建订阅请求
	falsePtr := false
	commitment := pb.CommitmentLevel_PROCESSED
	subscription := &pb.SubscribeRequest{
		Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
			"pump_subscription": {
				Vote:           &falsePtr,
				Failed:         &falsePtr,
				AccountInclude: []string{pumpAddr},
			},
		},
		Commitment: &commitment,
	}

	// 创建 gRPC 流
	stream, err := grpcClient.Subscribe(ctx)
	if err != nil {
		log.Fatalf("Failed to start subscription: %v", err)
	}

	// 发送订阅请求
	if err := stream.Send(subscription); err != nil {
		log.Fatalf("Failed to send subscription request: %v", err)
	}

	// 创建 ping 请求
	var globalID atomic.Int32
	pingRequest := &pb.SubscribeRequest{
		Ping: &pb.SubscribeRequestPing{
			Id: 0,
		},
	}
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	// 在单独的 goroutine 中发送 ping
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pingRequest.Ping.Id = globalID.Add(1)
				lastPingTime = time.Now() // 记录发送时间
				if err := stream.Send(pingRequest); err != nil {
					log.Printf("Failed to send ping: %v", err)
				}
			}
		}
	}()

	// 监控代币事件
	monitorTokens(ctx, stream)
}
