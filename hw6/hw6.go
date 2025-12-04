package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "id",
				Required: true,
			},
			&cli.IntFlag{Name: "n",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "root",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			id := cmd.Int("id")
			nNodes := cmd.Int("n")
			root := cmd.Int("root")

			if id == root {
				producer := newProducer(id, nNodes)
				producer.run()
			} else {
				consumer := newConsumer(id, root)
				consumer.run()
			}

			return nil
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

var basePort = 8000

type message struct {
	Id   int    `json:"id"`
	Data string `json:"data"`
}

func send(addr *net.UDPAddr, msg message) {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	bytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	conn.Write(bytes)
}

func receive(conn *net.UDPConn) *message {
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		panic(err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil
		}

		panic(err)
	}

	var msg message
	if err := json.Unmarshal(buffer[:n], &msg); err != nil {
		panic(err)
	}

	return &msg
}

type producer struct {
	id           int
	conn         *net.UDPConn
	consumerAddr []*net.UDPAddr
}

func newProducer(id, nNodes int) *producer {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%v", basePort+id))
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	consumerAddr := make([]*net.UDPAddr, 0, nNodes-1)
	for i := range nNodes {
		if i == id {
			continue
		}

		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%v", basePort+i))
		if err != nil {
			panic(err)
		}

		consumerAddr = append(consumerAddr, addr)
	}

	return &producer{
		id:           id,
		conn:         conn,
		consumerAddr: consumerAddr,
	}
}

func (p *producer) run() {
	defer p.conn.Close()
	msg_id := 0

	for {
		msg := message{
			Id:   msg_id,
			Data: "hello world",
		}
		msg_id++

		for _, addr := range p.consumerAddr {
			go func() {
				for range 5 {
					send(addr, msg)
					fmt.Printf("[%v] - sent message %v to %v:%v\n", p.id, msg, addr.IP, addr.Port)

					ack := receive(p.conn)
					if ack != nil {
						fmt.Printf("[%v] - %v:%v acknowledged message %v\n", p.id, addr.IP, addr.Port, msg)
						break
					}

					time.Sleep(500 * time.Millisecond)
				}
			}()
		}

		time.Sleep(3 * time.Second)
	}
}

type consumer struct {
	id           int
	conn         *net.UDPConn
	producerAddr *net.UDPAddr
}

func newConsumer(id, root int) *consumer {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%v", basePort+id))
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	producerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%v", basePort+root))
	if err != nil {
		panic(err)
	}

	return &consumer{
		id:           id,
		conn:         conn,
		producerAddr: producerAddr,
	}
}

func (c *consumer) run() {
	defer c.conn.Close()
	last_msg_id := -1

	for {
		msg := receive(c.conn)
		if msg != nil {
			if msg.Id > last_msg_id {
				last_msg_id = msg.Id
				fmt.Printf("[%v] - received message %v\n", c.id, msg)
			}

			send(c.producerAddr, *msg)
			fmt.Printf("[%v] - acknowledged message %v from %v:%v\n", c.id, msg, c.producerAddr.IP, c.producerAddr.Port)
		}
	}
}
