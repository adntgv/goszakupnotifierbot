package main

import (
	"log"
	"time"
)

func main() {
	go botRun()

	for {
		newAnnounces, err := getNewAnnounces()
		if err != nil {
			log.Println(err)
		}
 
		for _, announcement := range newAnnounces {
			sendMailOffer(announcement)
			sendToAllChats(announcement.String())
		} 

		time.Sleep(time.Minute)
	}
}

