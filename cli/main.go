package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/spf13/viper"

	. "github.com/mtrense/soil/config"
	"github.com/mtrense/soil/logging"
	"github.com/spf13/cobra"
	"github.com/ticker-es/client-go/client"
	"github.com/ticker-es/client-go/eventstream/base"
	"github.com/ticker-es/client-go/support"
)

var (
	version = "none"
	commit  = "none"
	app     = NewCommandline("ticker",
		Short("Run the ticker client"),
		Flag("connect", Str("localhost:6677"), Abbr("c"), Description("Server to connect to"), Mandatory(), Persistent(), Env()),
		Flag("token", Str(""), Abbr("a"), Description("Token to use for authentication against the Ticker Server"), Persistent(), Env()),
		FlagLogFile(),
		FlagLogFormat(),
		FlagLogLevel("warn"),
		SubCommand("emit",
			Short("Emit specified event"),
			Flag("topic", Str(""), Abbr("t"), Description("Select Topic and Type of the emitted event"), Persistent()),
			Flag("payload", Str("{}"), Abbr("p"), Description("The payload of the emitted event (- for stdin)"), Persistent()),
			Flag("from-stdin", Bool(), Description("Read events to be emitted from stdin"), Persistent()),
			Run(executeEmit),
		),
		SubCommand("sample",
			Short("Emit sample events"),
			Run(executeSample),
		),
		SubCommand("stream",
			Short("Stream a portion of the event stream"),
			Flag("format", Str("text"), Description("Format for Event output (text, json)"), Persistent()),
			Flag("omit-payload", Bool(), Description("Omit Payload in Event output"), Persistent()),
			Flag("pretty", Bool(), Description("Use pretty-mode in Event output"), Persistent()),
			Flag("selector", Str("/"), Abbr("s"), Description("Select which events to stream"), Persistent()),
			Flag("range", Str("1:"), Abbr("r"), Description("Select which events to stream"), Persistent()),
			Run(executeStream),
		),
		SubCommand("subscribe",
			Short("Subscribe to a specific event stream"),
			Flag("format", Str("text"), Description("Format for Event output (text, json)"), Persistent()),
			Flag("omit-payload", Bool(), Description("Omit Payload in Event output"), Persistent()),
			Flag("pretty", Bool(), Description("Use pretty-mode in Event output"), Persistent()),
			Flag("selector", Str("/"), Abbr("s"), Description("Select which events to subscribe to"), Persistent()),
			Flag("client-id", Str(""), Abbr("i"), Description("Unique Identifier for this subscription"), Mandatory(), Persistent(), Env()),
			Run(executeSubscribe),
		),
		SubCommand("metrics",
			Short("Show live metrics of the ticker server"),
			Run(executeMetrics),
		),
		Version(version, commit),
		Completion(),
	).GenerateCobra()
)

func init() {
	EnvironmentConfig("TICKER")
	logging.ConfigureDefaultLogging()
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

func executeEmit(cmd *cobra.Command, args []string) {
	ctx, _ := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	fromStdin, _ := cmd.Flags().GetBool("from-stdin")
	if cl, err := client.NewClient(viper.GetString("connect")); err == nil {
		if fromStdin {
			dec := json.NewDecoder(os.Stdin)
			for {
				var event base.Event
				err := dec.Decode(&event)
				if err == io.EOF {
					return
				}
				if err != nil {
					panic(err)
				}
				if _, err := cl.Emit(ctx, event); err != nil {
					panic(err)
				}
			}
		} else {
			payloadString, _ := cmd.Flags().GetString("payload")
			topicAndType, _ := cmd.Flags().GetString("topic")
			if selector, err := base.ParseSelector(topicAndType); err == nil {
				var payload map[string]interface{}
				if err := json.Unmarshal([]byte(payloadString), &payload); err != nil {
					panic(err)
				}
				event := base.Event{
					Aggregate:  selector.Aggregate,
					Type:       selector.Type,
					OccurredAt: time.Now(),
					Payload:    payload,
				}
				if _, err := cl.Emit(ctx, event); err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}
	} else {
		panic(err)
	}
}


func executeSample(cmd *cobra.Command, args []string) {

}

func executeStream(cmd *cobra.Command, args []string) {
	formatter := createFormatter(cmd)
	if cl, err := client.NewClient(viper.GetString("connect")); err == nil {
		ctx, _ := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
		count, err := cl.Stream(ctx, selectorFromFlags(cmd), bracketFromFlags(cmd), func(e *base.Event) error {
			return formatter(os.Stdout, e)
		})
		fmt.Printf("Handled %d events\n", count)
		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
}

func executeSubscribe(cmd *cobra.Command, args []string) {
	formatter := createFormatter(cmd)
	clientID := viper.GetString("client-id")
	if cl, err := client.NewClient(viper.GetString("connect")); err == nil {
		ctx, _ := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
		err := cl.Subscribe(ctx, clientID, selectorFromFlags(cmd), func(e *base.Event) error {
			return formatter(os.Stdout, e)
		})
		if err != nil {
			if ctx.Err() == context.Canceled {
				return
			}
			panic(err)
		}
	} else {
		panic(err)
	}
}

func executeMetrics(cmd *cobra.Command, args []string) {
	if cl, err := client.NewClient(viper.GetString("connect")); err == nil {
		ctx, _ := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
		for {
			cl.PrintServerState(ctx)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	} else {
		panic(err)
	}
}
