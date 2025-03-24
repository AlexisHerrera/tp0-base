package common

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	Apuesta        Apuesta
	BatchMaxAmount int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func readPacketData(ctx context.Context, conn net.Conn) ([]byte, error) {
	readCh := make(chan *Packet, 1)
	errCh := make(chan error, 1)

	// Runs concurrently
	go func() {
		// Usamos la función ReadPacket definida en packet.go
		packet, err := ReadPacket(conn)
		if err != nil {
			errCh <- err
			return
		}
		readCh <- packet
	}()

	select {
	case <-ctx.Done():
		// If context is cancelled, stops reading the buffer
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case packet := <-readCh:
		return packet.Data, nil
	}
}

func readBatchApuestas(scanner *bufio.Scanner, maxAmount int) ([]Apuesta, error) {
	var apuestas []Apuesta
	count := 0
	for count < maxAmount && scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		apuesta := Apuesta{
			Nombre:     strings.TrimSpace(fields[0]),
			Apellido:   strings.TrimSpace(fields[1]),
			Documento:  strings.TrimSpace(fields[2]),
			Nacimiento: strings.TrimSpace(fields[3]),
			Numero:     strings.TrimSpace(fields[4]),
		}
		apuestas = append(apuestas, apuesta)
		count++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return apuestas, nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context) {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	file, err := os.Open("agency.csv")
	if err != nil {
		log.Criticalf("action: open_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		select {
		case <-ctx.Done():
			// If context is cancelled, stop the loop
			log.Infof("action: loop_cancelled | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

		batch, err := readBatchApuestas(scanner, c.config.BatchMaxAmount)
		if err != nil {
			log.Errorf("action: read_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		if len(batch) == 0 {
			log.Infof("action: no_more_apuestas | result: finish | client_id: %v", c.config.ID)
			return
		}
		log.Infof("action: batch_read | result: success | client_id: %v | cantidad: %d", c.config.ID, len(batch))
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()
		apuestaToSend := batch[0]

		// Load config and serializes it
		msgData, err := SerializeApuesta(apuestaToSend)
		if err != nil {
			log.Errorf("action: serialize_apuesta | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			return
		}
		packet := NewPacket(msgData)
		log.Info("action: serialize_apuesta | result: success")

		// Writes every byte, fails otherwise
		if err := packet.Write(c.conn); err != nil {
			log.Errorf("action: send_apuesta | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			return
		}
		log.Info("action: send_apuesta | result: success")

		// Handles context cancel, error and successful reads.
		data, err := readPacketData(ctx, c.conn)
		log.Infof("action: socket_closed | result: success | client_id: %v", c.config.ID)
		c.conn.Close()

		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Infof("action: receive_message | result: cancelled | client_id: %v", c.config.ID)
				return
			}
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			string(data),
		)
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", c.config.Apuesta.Documento, c.config.Apuesta.Numero)
		// Wait a time between sending one message and the next one
		select {
		case <-ctx.Done():
			// If context is cancelled, stop the sleep
			log.Infof("action: loop_cancelled_during_sleep | result: success | client_id: %v", c.config.ID)
			return
		case <-time.After(c.config.LoopPeriod):
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
