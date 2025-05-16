package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)
func main(){
	

	app := fiber.New()
	app.Get("/",func(c *fiber.Ctx)error{
		output := make(map[int] string)

		output[1]="bhargav"
		output[2]="srya"
		c.Status(200)
		c.JSON(output)
		return nil
	})	
	err:=app.Listen(":8100")
	if(err!=nil){
		fmt.Println("Error in Starting server: ",err)
	}
}



