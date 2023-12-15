package main

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Message struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type Data struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
}

var items []Data

func main() {
	app := fiber.New()

	app.Static("/", "./public")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./public/index.html")
	})

	// WebSocket endpoint
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				break
			}
			var message Message
			err = json.Unmarshal(msg, &message)
			if err != nil {
				log.Println("JSON unmarshal error:", err)
				break
			}
			switch message.Action {
			case "create":
				var newData Data
				dataStr, ok := message.Data.(string)
				if !ok {
					log.Println("Invalid data format received for create action")
					break
				}
				if err := json.Unmarshal([]byte(dataStr), &newData); err != nil {
					log.Println("Error unmarshaling data for create action:", err)
					break
				}
				items = append(items, newData)
				broadcastItems(c, Message{Action: "read"})
			case "read":
				sendItems(c)
			case "update":
				var updateData Data
				dataStr, ok := message.Data.(string)
				if !ok {
					log.Println("Invalid data format received for update action")
					break
				}
				if err := json.Unmarshal([]byte(dataStr), &updateData); err != nil {
					log.Println("Error unmarshaling data for update action:", err)
					break
				}
				updateItem(updateData)
				broadcastItems(c, Message{Action: "read"})
			case "delete":
				var deleteData Data
				dataStr, ok := message.Data.(string)
				if !ok {
					log.Println("Invalid data format received for delete action")
					break
				}
				if err := json.Unmarshal([]byte(dataStr), &deleteData); err != nil {
					log.Println("Error unmarshaling data for delete action:", err)
					break
				}
				deleteItem(deleteData.ID)
				broadcastItems(c, Message{Action: "read"})

			}
		}
	}))

	log.Fatal(app.Listen(":3000"))
}

func sendItems(c *websocket.Conn) {
	response := Message{
		Action: "read",
		Data:   items,
	}
	responseBytes, _ := json.Marshal(response)
	c.WriteMessage(websocket.TextMessage, responseBytes)
}

func updateItem(updatedItem Data) {
	for index, item := range items {
		if item.ID == updatedItem.ID {
			items[index] = updatedItem
			break
		}
	}
}

func deleteItem(id int) {
	for index, item := range items {
		if item.ID == id {
			items = append(items[:index], items[index+1:]...)
			break
		}
	}
}

func broadcastItems(c *websocket.Conn, message Message) {
	responseBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	c.WriteMessage(websocket.TextMessage, responseBytes)
}
