/*
The event package implements events, an event source and an event receiver.
You should embed the event.Source and event.Receiver instances in your
instance in order to provide facilities for subscribing to events
and receiving events. For example,

	type Task struct {
		event.Source
		event.Receiver
	}

	func main() {
		a,b  := new(Task), new(Task)

		// Receive events from task b in task a
		go a.Recv(context.TODO(), process, b)

		// Emit an event in task b which is processed in task a
		evt := event.New(context.TODO(), "key", "value")
		evt.Emit(b.C())
	}

*/
package event
