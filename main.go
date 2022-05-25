package main

import (
	"log"
	"regexp"

	"github.com/rjeczalik/notify"
)

type renameEvent struct {
	before struct {
		name      string
		compliant bool
		event     notify.EventInfo
	}
	after struct {
		name      string
		compliant bool
		event     notify.EventInfo
	}
}

func newRenameEvent(name string, event notify.EventInfo) *renameEvent {
	re := renameEvent{}
	return &re
}

func main() {
	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	events := make(chan notify.EventInfo, 1)

	// Set up a watchpoint listening on events within current working directory.
	// Dispatch each create and remove events separately to c.
	if err := notify.Watch(".", events, notify.Rename); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(events)

	/**
	Try to pair sequential rename events as
	...
	after,
	before
	...
	-> {
		before,
		after
	}
	No idea if the emmited events are actually in order...
	@todo maybe verify that??
	*/
	//eventPairs := make(map[string]renameEvent)
	temp := renameEvent{}
	for event := range events {
		ok, filename := extractFilename(event.Path())
		if !ok {
			log.Fatal("Path not ok! - " + event.Path())
		}
		ok, _ = isNameCompliant(filename)
		if temp.after.name == "" {
			temp.after.name = filename
			temp.after.compliant = ok
			temp.after.event = event
		} else if temp.before.name == "" {
			temp.before.name = filename
			temp.before.compliant = ok
			temp.before.event = event
			log.Println("Event details: ", temp.before.name, "-->", temp.after.name)
			log.Println("isNameCompliant? ", temp.before.compliant, "-->", temp.after.compliant)
			temp = renameEvent{}
		}
	}

}

func extractFilename(path string) (ok bool, filename string) {
	filenameRegExp := regexp.MustCompile(`(.+?)(?:\\|/)([\w-\. ]+)$`)
	match := filenameRegExp.FindStringSubmatch(path)
	if match == nil {
		return false, ""
	}
	return true, match[2]
}

func isNameCompliant(filename string) (ok bool, details map[string]string) {
	validNameExp := regexp.MustCompile(`.*?(?P<jobNumber>\d+)_(?P<client>[a-zA-Z]+)_(?P<name>[a-zA-Z]+)(?P<accronym>-[A-Z]{2,4})?(?P<dimensions>-\d{1,4}.?\d{1,4}x\d{1,4}.?\d{1,4})?(?P<variant>-[A-Z])?(?P<version>-(?:[vV]\d{1,2}|\d{6}))?\.[a-z]+`)

	match := validNameExp.FindStringSubmatch(filename)
	if match == nil {
		return false, nil
	}
	result := make(map[string]string)
	for i, name := range validNameExp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return true, result
}
