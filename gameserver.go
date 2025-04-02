package fakegameserver

import (
	"context"
	"reflect"
	"time"

	"github.com/antiphp/fakegameserver/internal/queue"
	"github.com/google/uuid"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
)

// Message is a message produces by a message producer and consumable by a message consumer.
type Message struct {
	ID          string
	Type        MessageType
	Description string
	Error       error
	Payload     any
	Origin      string
	Created     time.Time
}

// MessageType is the type of a message.
type MessageType string

const (
	// MessageTypeExit is the message type for exit messages.
	MessageTypeExit MessageType = "exit"

	// MessageTypeInfo is the message type for info messages.
	MessageTypeInfo MessageType = "info"
)

// Queue is a message queue.
type Queue interface {
	Add(Message)
}

// Producer is a message producer.
type Producer interface {
	Run(context.Context, Queue)
}

// Consumer is a message consumer.
type Consumer interface {
	Consume(Message)
}

// GameServer is the game server.
type GameServer struct {
	queue     *queue.Fifo[Message]
	producers []Producer
	consumers []Consumer

	log *logger.Logger
}

// New creates a new game server.
func New(log *logger.Logger) *GameServer {
	return &GameServer{
		queue: queue.NewFifo[Message](),
		log:   log,
	}
}

// AddProducer adds a message producer to the game server.
func (g *GameServer) AddProducer(p ...Producer) {
	g.producers = append(g.producers, p...)
}

// AddConsumer adds a mesage consumer to the game server.
func (g *GameServer) AddConsumer(c ...Consumer) {
	g.consumers = append(g.consumers, c...)
}

// AddHandler adds a message handler to the game server.
func (g *GameServer) AddHandler(hdlr any) {
	p, ok1 := hdlr.(Producer)
	if ok1 {
		g.AddProducer(p)
	}
	c, ok2 := hdlr.(Consumer)
	if ok2 {
		g.AddConsumer(c)
	}
	if !ok1 && !ok2 {
		panic("Handler must either implement Producer or Consumer interface") // Developer error.
	}
}

// Run starts the game server and runs all producers and consumers.
func (g *GameServer) Run(ctx context.Context) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	context.AfterFunc(ctx, func() {
		g.queue.Shutdown()
	})

	for _, p := range g.producers {
		name := reflect.TypeOf(p).Elem().String()

		go p.Run(ctx, queueFn(func(m Message) {
			m.ID = uuid.NewString()
			m.Created = time.Now()
			m.Origin = name
			g.queue.Add(m)
		}))
	}

	for {
		msg, shutdown := g.queue.Get()
		if shutdown {
			return "", nil
		}

		log := g.log.With(lctx.Str("desc", msg.Description), lctx.Str("type", string(msg.Type)))
		if msg.Error != nil {
			log = log.With(lctx.Err(msg.Error))
		}
		log.Info("Game server message received")

		for _, c := range g.consumers {
			c.Consume(msg)
		}

		if msg.Type == MessageTypeExit {
			return msg.Description, msg.Error
		}
	}
}

type queueFn func(Message)

func (fn queueFn) Add(m Message) {
	fn(m)
}
