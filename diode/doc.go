// Package diode implements a ring buffer that's optimized for high-throughput
// scenarios where losing data is acceptable. The ring buffer does its best to
// not "push back" on the producer.
package diode
