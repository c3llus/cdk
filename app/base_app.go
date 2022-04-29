package app

import (
	"context"
	"time"

	"github.com/c3llus/cdk/app/resource"
	"github.com/c3llus/cdk/hodler"
)

type BaseApp struct {
	resource resource.Resource
	// TODO:
	// 1. add cfg
	// 2. add nsq
}

func newBaseApp() (*BaseApp, error) {
	// TODO:
	// 1. add cfg param

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	hdlr, err := hodler.New(ctx)
	if err != nil {
		return nil, err
	}

	return &BaseApp{
		resource: hdlr,
	}, nil
}
