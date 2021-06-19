package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ticker-es/client-go/client"
	"github.com/ticker-es/client-go/eventstream/base"
)

func selectorFromFlags(cmd *cobra.Command) *base.Selector {
	sel, _ := cmd.Flags().GetString("selector")
	if selector, err := base.ParseSelector(sel); err == nil {
		return selector
	}
	return nil
}

func bracketFromFlags(cmd *cobra.Command) *base.Bracket {
	rang, err := cmd.Flags().GetString("range")
	if err != nil {
		panic(err)
	}
	ranges := strings.Split(rang, ":")
	first, err := strconv.ParseInt(ranges[0], 10, 64)
	if err != nil {
		first = 0
	}
	last, err := strconv.ParseInt(ranges[1], 10, 64)
	if err != nil {
		last = -1
	}
	return &base.Bracket{
		NextSequence: first,
		LastSequence: last,
	}
}

func createFormatter(cmd *cobra.Command) client.Formatter {
	format, _ := cmd.Flags().GetString("format")
	omitPayload, _ := cmd.Flags().GetBool("omit-payload")
	pretty, _ := cmd.Flags().GetBool("pretty")
	var formatter client.Formatter
	switch strings.ToLower(format) {
	case "json":
		formatter = client.JsonFormatter(pretty)
	default:
		formatter = client.TextFormatter(pretty)
	}
	if omitPayload {
		formatter = client.OmitPayload(formatter)
	}
	return formatter
}

func loadEvents(files ...string) []base.Event {
	var events []base.Event
	for _, arg := range files {
		if data, err := ioutil.ReadFile(arg); err == nil {
			var ev []base.Event
			if err := json.Unmarshal(data, &ev); err != nil {
				panic(err)
			}
			events = append(events, ev...)
		} else {
			panic(err)
		}
	}
	return events
}
