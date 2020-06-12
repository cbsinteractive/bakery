package logging

import (
	"context"

	"github.com/rs/zerolog"
)

type Params map[string]interface{}

func UpdateCtx(ctx context.Context, p Params) {
	log := zerolog.Ctx(ctx)
	log.UpdateContext(func(c zerolog.Context) zerolog.Context {
		for k, v := range p {
			c = c.Interface(k, v)
		}
		return c
	})
}
