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

type Message struct {
	ID          string
	Type        MessageType
	Description string
	Error       error
	Payload     any
	Origin      string
	Created     time.Time
}

type MessageType string

const (
	MessageTypeExit MessageType = "exit"
	MessageTypeInfo MessageType = "info"
)

type Queue interface {
	Add(Message)
}

type Producer interface {
	Run(context.Context, Queue)
}

type Consumer interface {
	Consume(Message) Message // TODO: remove returning the message.
}

type GameServer struct {
	queue     *queue.Fifo[Message]
	producers []Producer
	consumers []Consumer

	log *logger.Logger
}

func New(log *logger.Logger) *GameServer {
	return &GameServer{
		queue: queue.NewFifo[Message](),
		log:   log,
	}
}

func (g *GameServer) AddProducer(p ...Producer) {
	g.producers = append(g.producers, p...)
}

func (g *GameServer) AddConsumer(c ...Consumer) {
	g.consumers = append(g.consumers, c...)
}

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
			msg = c.Consume(msg)
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
