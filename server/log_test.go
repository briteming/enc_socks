package server

import "testing"
import log "github.com/sirupsen/logrus"

func TestLog(t *testing.T) {
    log.Errorf("Hello test, %d", 1)
    log.Printf("haha, %s", "xxx")
}

