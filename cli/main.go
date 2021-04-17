package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"

	. "github.com/mtrense/soil/config"
	"github.com/mtrense/soil/logging"
	"github.com/spf13/cobra"
	"github.com/ticker-es/client-go/client"
	"github.com/ticker-es/client-go/eventstream/base"
	"github.com/ticker-es/client-go/rpc"
	"github.com/ticker-es/client-go/support"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	version = "none"
	commit  = "none"
	app     = NewCommandline("ticker",
		Short("Run the ticker client"),
		Flag("connect", Str("localhost:6677"), Description("Server to connect to"), Mandatory(), Persistent(), Env()),
		SubCommand("emit",
			Short("Emit specified event"),
			Flag("topic", Str(""), Abbr("t"), Description("Select Topic and Type of the emitted event"), Persistent()),
			Flag("payload", Str("{}"), Abbr("p"), Description("The payload of the emitted event (- for stdin)"), Persistent()),
			Run(executeClientEmit),
		),
		SubCommand("sample",
			Short("Emit sample events"),
			Run(executeClientSample),
		),
		SubCommand("stream",
			Short("Stream a portion of the event stream"),
			Flag("format", Str("text"), Description("Format for Event output (text, json)"), Persistent()),
			Flag("omit-payload", Bool(), Description("Omit Payload in Event output"), Persistent()),
			Flag("pretty", Bool(), Description("Use pretty-mode in Event output"), Persistent()),
			Flag("selector", Str(">/*"), Abbr("s"), Description("Select which events to stream"), Persistent()),
			Flag("range", Str("1:"), Abbr("r"), Description("Select which events to stream"), Persistent()),
			Run(executeClientStream),
		),
		SubCommand("subscribe",
			Short("Subscribe to a specific event stream"),
			Flag("format", Str("text"), Description("Format for Event output (text, json)"), Persistent()),
			Flag("omit-payload", Bool(), Description("Omit Payload in Event output"), Persistent()),
			Flag("pretty", Bool(), Description("Use pretty-mode in Event output"), Persistent()),
			Flag("selector", Str(">/*"), Description("Select which events to subscribe to"), Persistent()),
			Run(executeClientSubscribe),
		),
		SubCommand("metrics",
			Short("Show live metrics of the ticker server"),
			Run(executeClientMetrics),
		),
		Version(version, commit),
		Completion(),
	).GenerateCobra()
)

func init() {
	EnvironmentConfig("TICKER")
	ApplyLogFlags(app)
	logging.ConfigureDefaultLogging()
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

func executeClientEmit(cmd *cobra.Command, args []string) {
	payloadString, _ := cmd.Flags().GetString("payload")
	topicAndType, _ := cmd.Flags().GetString("topic")
	if selector, err := base.ParseSelector(topicAndType); err == nil {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(payloadString), &payload); err != nil {
			panic(err)
		}
		cl := client.NewClient(clientConnect())
		ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
		event := base.Event{
			Aggregate:  selector.Aggregate,
			Type:       selector.Type,
			OccurredAt: time.Now(),
			Payload:    payload,
		}
		if _, err := cl.Emit(ctx, event); err != nil {
			panic(err)
		} else {

		}
	} else {
		panic(err)
	}
}

func executeClientSample(cmd *cobra.Command, args []string) {

}

func executeClientStream(cmd *cobra.Command, args []string) {
	formatter := createFormatter(cmd)
	cl := client.NewClient(clientConnect())
	ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	err := cl.Stream(ctx, selectorFromFlags(cmd), bracketFromFlags(cmd), func(e *base.Event) error {
		return formatter(os.Stdout, e)
	})
	if err != nil {
		panic(err)
	}
}

func executeClientSubscribe(cmd *cobra.Command, args []string) {
	//formatter := createFormatter(cmd)
	cl := client.NewClient(clientConnect())
	ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	cl.Subscribe(ctx)
}

func executeClientMetrics(cmd *cobra.Command, args []string) {
	conn := clientConnect()
	admin := rpc.NewMaintenanceClient(conn)
	ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	for {
		if state, err := admin.GetServerState(ctx, &emptypb.Empty{}); err == nil {
			fmt.Printf("uptime: %5ds   |   active connections: %3d   |   events stored: %8d\n", state.Uptime, state.ConnectionCount, state.EventCount)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}
