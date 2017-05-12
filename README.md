# diodes

Diodes are ring buffers manipulated via atomics.

Diodes are optimized for high throughput scenarios where losing data is
acceptable. Unlike a channel, a diode will overwrite data in lieu of blocking.
A diode does its best to not "push back" on the producer. In other words,
invoking `Set()` on a diode never blocks.

### Example
```go

d := diodes.NewPoller(diodes.NewOneToOne(b.N, diodes.AlertFunc(func(missed int) {
	log.Printf("Dropped %d messages", missed)
})))

go func() {
	for i := 0; i < 1000; i++ {
		// Warning: Do not use i. By taking the address,
		// you would not get each value
		j := i 		d.Set(diodes.GenericDataType(&data))
		d.Set(diodes.GenericDataType(&j))
	}
}()

for {
	d.Next()
}

```

### Dropping Data

The diode takes an `Alerter` as an argument to alert the user code to when
the read noticed it missed data. It is important to note that the go-routine
consuming from the diode is used to signal the alert.

When the diode notices it has fallen behind, it will move the read index to
the new write index and therefore drop more than a single message.

There are two things to consider when choosing a diode:
1. Storage layer
2. Access layer

### Storage Layer

##### OneToOne

The OneToOne diode is meant to be used by one producing (invoking `Set()`)
go-routine and a (different) consuming (invoking `TryNext()`) go-routine. It
is not thread safe for multiple readers or writers.

##### ManyToOne

The ManyToOne diode is optimized for many producing (invoking `Set()`)
go-routines and a single consuming (invoking `TryNext()`) go-routine. It is
not thread safe for multiple readers.

It is recommended to have a larger diode buffer size if the number of producers
is high. This is to avoid the diode from having to mitigate write collisions
(it will call its alert function if this occurs).

### Access Layer

##### Poller

The Poller uses polling via `time.Sleep(...)` when `Next()` is invoked. While
polling might seem sub-optimal, it allows the producer to be completely
decoupled from the consumer. If you require very minimal push back on the
producer, then the Poller is a better choice. However, if you require several
diodes (e.g. one per connected client), then having several go-routines
polling (sleeping) may be hard on the scheduler.

##### Waiter

The Waiter uses a conditional mutex to manage when the reader is alerted
of new data. While this method is great for the scheduler, it does have
extra overhead for the producer. Therefore, it is better suited for situations
where you have several diodes and can afford slightly slower producers.


### Benchmarks

There are benchmarks that compare the various storage and access layers to
channels. To run them:

```
go test -bench=. -run=NoTest
```
