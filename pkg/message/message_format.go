package message

import (
	"os"
	"time"

	"github.com/fatih/color"
)

type MyTime struct {
	CurrentTime time.Time
	Ack         bool
}

func MyError(err error, message string) {
	color.Red("******************** Error ********************")
	color.Red("Error happened in > %s", message)
	color.Red("Error Message > %s", err.Error())
	color.Red("***********************************************")
	os.Exit(1)
}

func MyWriteMessage(m any) {
	color.Green("[WRITE] %v", m)
}

func MyReadMessage(m any) {
	color.Green("[READ ] %v", m)
}

func MyMessage(s string) {
	color.Blue(s)
}
