package main

import "archi/setting"

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	setting.InitViper()

	app := InitApp()
	server := app.engine
	err := server.Run(":8080")
	if err != nil {
		return
	}
}
