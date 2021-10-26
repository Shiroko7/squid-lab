package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func registerDeath(msg string) {
	f, err := os.OpenFile("pozo.txt", os.O_RDWR|os.O_CREATE, 0600)
	failOnError(err, "Failed to create text file")

	reader := bufio.NewReader(f)
	line := "EMPTY"
	for {
		line_i, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else {
			line = line_i
		}
		failOnError(err, "Failed to read line on text file")
		log.Println(line)
	}

	if line == "EMPTY" {
		_, err = f.WriteString(msg + " 100000000\n")
	} else {
		line_fields := strings.Fields(line)
		string_price := line_fields[len(line_fields)-1]
		int_price, err := strconv.Atoi(string_price)
		failOnError(err, "Failed to convert string value to int")
		int_price = int_price + 100000000
		string_price = strconv.Itoa(int_price)
		_, err = f.WriteString(msg + " " + string_price + "\n")
	}
	failOnError(err, "Failed to write on text file")

	err = f.Close()
	failOnError(err, "Failed to close text file")
	log.Println("Successfully saved player's death")
}

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Craete the queue
	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			msg := string(d.Body)
			registerDeath(msg)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
