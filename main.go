package main

import (
	"archi/setting"
	"context"
	"time"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	setting.InitViper()
	setting.InitValidate()
	setting.InitPrometheus()
	tpCancel := setting.InitOTEL()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		tpCancel(ctx)
	}()

	app := InitApp()

	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	app.cron.Start()
	defer func() {
		<-app.cron.Stop().Done()
	}()

	server := app.engine
	if err := server.Run(":8080"); err != nil {
		return
	}
}
