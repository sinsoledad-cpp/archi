package main

import "archi/setting"

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	setting.InitViper()

	app := InitApp()

	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	//app.cron.Start()
	//defer func() {
	//	<-app.cron.Stop().Done()
	//}()

	server := app.engine
	err := server.Run(":8080")
	if err != nil {
		return
	}
}
