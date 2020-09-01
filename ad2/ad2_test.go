package ad2_test

import (
	"fmt"
	"go-measure/ad2"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAd2(t *testing.T) {
	a, err := ad2.New()
	assert.NoError(t, err)
	a.Open()
	_ = a.SetOutput(0, 3.0, 0.0)
	_ = a.SetOutput(1, -2.0, 0.0)
	fmt.Printf("Done\n")
	a.Close()
}
