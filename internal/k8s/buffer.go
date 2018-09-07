package k8s

import (
	"github.com/mad01/totem/internal/workgroup"
	"github.com/sirupsen/logrus"

	"k8s.io/client-go/tools/cache"
)

type buffer struct {
	ev chan interface{}
	logrus.StdLogger
	rh cache.ResourceEventHandler
}

type addEvent struct {
	obj interface{}
}

type updateEvent struct {
	oldObj, newObj interface{}
}

type deleteEvent struct {
	obj interface{}
}

// NewBuffer returns a ResourceEventHandler which buffers and serialises ResourceEventHandler events.
func NewBuffer(g *workgroup.Group, rh cache.ResourceEventHandler, log logrus.FieldLogger, size int) cache.ResourceEventHandler {
	buf := &buffer{
		ev:        make(chan interface{}, size),
		StdLogger: log.WithField("context", "buffer"),
		rh:        rh,
	}
	g.Add(buf.loop)
	return buf
}

func (b *buffer) loop(stop <-chan struct{}) {
	b.Println("started")
	defer b.Println("stopped")

	for {
		select {
		case ev := <-b.ev:
			switch ev := ev.(type) {
			case *addEvent:
				b.rh.OnAdd(ev.obj)
			case *updateEvent:
				b.rh.OnUpdate(ev.oldObj, ev.newObj)
			case *deleteEvent:
				b.rh.OnDelete(ev.obj)
			default:
				b.Printf("unhandled event type: %T: %v", ev, ev)
			}
		case <-stop:
			return
		}
	}
}

func (b *buffer) OnAdd(obj interface{}) {
	b.send(&addEvent{obj})
}

func (b *buffer) OnUpdate(oldObj, newObj interface{}) {
	b.send(&updateEvent{oldObj, newObj})
}

func (b *buffer) OnDelete(obj interface{}) {
	b.send(&deleteEvent{obj})
}

func (b *buffer) send(ev interface{}) {
	select {
	case b.ev <- ev:
		// all good
	default:
		b.Printf("event channel is full, len: %v, cap: %v", len(b.ev), cap(b.ev))
		b.ev <- ev
	}
}
