package common

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const maxPayloadSize = 8*1024 - 4

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

func readBatchApuestasBytes(scanner *bufio.Scanner, maxAmount int, leftover *string) ([]byte, int, error) {
	var payload []byte
	count := 0
	// If there is something leftover, then process it first
	if *leftover != "" {
		line := *leftover
		*leftover = ""
		apuesta, err := ApuestaFromCSVLine(line)
		if err != nil {
			return nil, 0, err
		}
		serialized, err := SerializeApuesta(apuesta)
		if err != nil {
			return nil, 0, err
		}
		subpacket := NewPacket(serialized)
		subPacketBytes := subpacket.Bytes()
		// There is no need to check if the payload is too big, as it was already checked before
		payload = append(payload, subPacketBytes...)
		count++
	}

	for count < maxAmount && scanner.Scan() {
		line := scanner.Text()
		apuesta, err := ApuestaFromCSVLine(line)
		if err != nil {
			return nil, 0, err
		}
		serialized, err := SerializeApuesta(apuesta)
		if err != nil {
			return nil, count, err
		}
		if len(serialized) > maxPayloadSize {
			log.Errorf("action: serialize_apuesta | result: skip | error: apuesta exceeds max payload size: %v", line)
			continue
		}
		subpacket := NewPacket(serialized)
		subPacketBytes := subpacket.Bytes()
		// Checks if the payload is too big
		if len(payload)+len(subPacketBytes) <= maxPayloadSize {
			payload = append(payload, subPacketBytes...)
			count++
		} else {
			*leftover = line
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, count, err
	}
	return payload, count, nil
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
	var leftover string
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		select {
		case <-ctx.Done():
			// If context is cancelled, stop the loop
			log.Infof("action: loop_cancelled | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

		msgData, batchCount, err := readBatchApuestasBytes(scanner, c.config.BatchMaxAmount, &leftover)
		if err != nil {
			log.Errorf("action: read_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		if batchCount == 0 {
			return
		}
		log.Infof("action: batch_read | result: success | client_id: %v | bets_sent: %d", c.config.ID, batchCount)
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		packet := NewPacket(msgData)
		log.Infof("action: Packet created | result: success | size: %v", len(packet.Data))

		// Writes every byte, fails otherwise
		if err := packet.Write(c.conn); err != nil {
			log.Errorf("action: send_apuesta | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			return
		}
		log.Info("action: batch_sent | result: success")

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
