package common

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	Apuesta       Apuesta
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

func readLineWithContext(ctx context.Context, conn net.Conn, bufferSize int) (string, error) {
	readCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Runs concurrently
	go func() {
		buffer := make([]byte, bufferSize)
		var message []byte

		for {
			n, err := conn.Read(buffer)
			if n > 0 {
				chunk := buffer[:n]
				message = append(message, chunk...)

				// If chunk contains newline, stops reading
				if bytes.Contains(chunk, []byte{'\n'}) {
					readCh <- string(message)
					return
				}
			}

			if err != nil {
				if err == io.EOF {
					readCh <- string(message)
					return
				}

				errCh <- err
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		// If context is cancelled, stops reading the buffer
		return "", ctx.Err()
	case err := <-errCh:
		return "", err
	case msg := <-readCh:
		return msg, nil
	}
}

func writeFull(conn net.Conn, data []byte) error {
	total := 0
	for total < len(data) {
		n, err := conn.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context) {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		select {
		case <-ctx.Done():
			// If context is cancelled, stop the loop
			log.Infof("action: loop_cancelled | result: success | client_id: %v", c.config.ID)
			return
		default:
		}
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		// Load config and serializes it
		msgBytes, err := SerializeApuestaWithLength(c.config.Apuesta)
		if err != nil {
			log.Errorf("action: serialize_apuesta | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			return
		}
		log.Info("action: serialize_apuesta | result: success")

		// Writes every byte, fails otherwise
		if err := writeFull(c.conn, msgBytes); err != nil {
			log.Errorf("action: send_apuesta | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			return
		}
		log.Info("action: send_apuesta | result: success")

		// Handles context cancel, error and successful reads.
		msg, err := readLineWithContext(ctx, c.conn, 1024)
		c.conn.Close()
		log.Infof("action: socket_closed | result: success | client_id: %v", c.config.ID)

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
			msg,
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
