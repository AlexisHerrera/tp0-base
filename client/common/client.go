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
		return err
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

func (c *Client) FinalizeAndQueryWinners(ctx context.Context) {
	delay := 2 * time.Second

	for {
		// Si se recibe SIGTERM se termina.
		select {
		case <-ctx.Done():
			log.Infof("action: consulta_ganadores | result: cancelled | client_id: %v", c.config.ID)
			return
		default:
		}

		if err := c.createClientSocket(); err != nil {
			log.Criticalf("action: create_socket | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		consulta := NewConsultaMessage(c.config.ID)
		if err := consulta.Write(c.conn); err != nil {
			log.Errorf("action: send_consulta | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			return
		}
		log.Infof("action: send_consulta | result: success | client_id: %v", c.config.ID)

		msg, err := ReadMessage(c.conn)
		c.conn.Close()
		if err != nil {
			log.Errorf("action: read_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		baseMsg, ok := msg.(*BaseMessage)
		if !ok || (baseMsg.MsgType != MsgTypeRespuestaWinner && baseMsg.MsgType != MsgTypeRespuestaWait) {
			log.Errorf("action: read_message | result: fail | client_id: %v | unexpected msg type: %v", c.config.ID, baseMsg.MsgType)
			return
		}

		// Si es de tipo MsgTypeRespuestaWait, se espera y se vuelve a consultar
		if baseMsg.MsgType == MsgTypeRespuestaWait {
			log.Infof("action: consulta_ganadores | result: in_progress | client_id: %v | waiting: %v", c.config.ID, delay)
			time.Sleep(delay)
			delay *= 2
			continue
		}

		winners, err := ParseRespuestaPayload(baseMsg.Payload)
		if err != nil {
			log.Errorf("action: parse_respuesta | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d | client_id: %v", len(winners), c.config.ID)
		return
	}
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
	for {
		select {
		case <-ctx.Done():
			// If context is cancelled, stop the loop
			log.Infof("action: loop_cancelled | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

		batch, err := ReadBatch(scanner, c.config, &leftover)
		if err != nil {
			log.Errorf("action: read_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		if batch.GetCount() == 0 {
			break
		}
		log.Infof("action: batch_read | result: success | client_id: %v | bets_sent: %d", c.config.ID, batch.GetCount())
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		log.Infof("action: Packet created | result: success | payload size: %v", batch.GetFinalPayloadInBytes())

		// Writes every byte, fails otherwise
		if err := batch.Write(c.conn); err != nil {
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
	c.FinalizeAndQueryWinners(ctx)
}
