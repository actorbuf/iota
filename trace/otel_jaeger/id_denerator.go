package otel_jaeger

import (
	crand "crypto/rand"
	"encoding/binary"
	otelt "go.opentelemetry.io/otel/trace"
	"math/rand"
	"sync"
)

var IDGenerator = defaultIDGenerator()
var idLock sync.Mutex

func NewSpanID() otelt.SpanID {
	idLock.Lock()
	defer idLock.Unlock()

	sid := otelt.SpanID{}
	IDGenerator.Read(sid[:])
	return sid
}

func defaultIDGenerator() *rand.Rand {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	return rand.New(rand.NewSource(rngSeed))
}

func NewTraceID() otelt.TraceID {
	idLock.Lock()
	defer idLock.Unlock()

	tid := otelt.TraceID{}
	IDGenerator.Read(tid[:])
	return tid
}
