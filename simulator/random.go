package simulator

import (
	"math"
	"math/rand"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func RandomMsgID() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(math.MaxUint16)
}

func RandomUUID() string {
	full := uuid.NewV4().String()
	ss := strings.Split(full, "-")
	short := ss[0]
	return short
}
