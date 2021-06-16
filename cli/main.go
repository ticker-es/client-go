package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"

	"google.golang.org/grpc/credentials"

	"github.com/eiannone/keyboard"

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
		Flag("insecure", Bool(), Description("Connect insecurely (without TLS/checking)"), Persistent(), Env()),
		Flag("ca-cert", Str(""), Description("CA certificate used to verify server connection"), Persistent(), EnvName("ca_cert")),
		Flag("client-cert", Str(""), Description("Client certificate"), Persistent(), EnvName("client_cert")),
		Flag("client-key", Str(""), Description("Client key"), Persistent(), EnvName("client_key")),
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
		SubCommand("play",
			Short("Play events from a file"),
			Flag("pause", Int(0), Abbr("p"), Description("Pause between emitting events (milliseconds)")),
			Flag("random", Bool(), Description("Randomize interval between 0 and pause")),
			Flag("manual", Bool(), Description("Advance manually to the next event")),
			Flag("sunflower", Bool(), Description("Use a sunflower 🌻 as progress symbol")),
			Args(cobra.MinimumNArgs(1)),
			Run(executePlay),
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
	cl := connect()
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
}

func executePlay(cmd *cobra.Command, args []string) {
	ctx, cancel := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	pause, _ := cmd.Flags().GetInt("pause")
	random, _ := cmd.Flags().GetBool("random")
	manual, _ := cmd.Flags().GetBool("manual")
	sunflower, _ := cmd.Flags().GetBool("sunflower")
	cl := connect()
	for _, arg := range args {
		if data, err := ioutil.ReadFile(arg); err == nil {
			var events []base.Event
			if err := json.Unmarshal(data, &events); err == nil {
				fmt.Printf("Processing %d events.\n", len(events))
				if manual {
					fmt.Print("Press any key to send the next event (q for quit) [")
				}
				for _, event := range events {
					if manual {
						if _, key, err := keyboard.GetSingleKey(); err == nil {
							if key == 0x03 || key == 'q' {
								cancel()
								return
							}
						} else {
							panic(err)
						}
						if sunflower {
							fmt.Print("🌻")
						} else {
							fmt.Print("•")
						}
					}
					if _, err := cl.Emit(ctx, event); err != nil {
						panic(err)
					}
					if pause > 0 {
						if random && pause > 1 {
							time.Sleep(time.Millisecond * time.Duration(rand.Intn(pause-1)+1))
						} else {
							time.Sleep(time.Millisecond * time.Duration(pause))
						}
					}
				}
				if manual {
					fmt.Println("]")
				}
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}

func executeSample(cmd *cobra.Command, args []string) {

}

func executeStream(cmd *cobra.Command, args []string) {
	formatter := createFormatter(cmd)
	cl := connect()
	ctx, _ := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	count, err := cl.Stream(ctx, selectorFromFlags(cmd), bracketFromFlags(cmd), func(e *base.Event) error {
		return formatter(os.Stdout, e)
	})
	fmt.Printf("Handled %d events\n", count)
	if err != nil {
		panic(err)
	}
}

func executeSubscribe(cmd *cobra.Command, args []string) {
	formatter := createFormatter(cmd)
	clientID := viper.GetString("client-id")
	cl := connect()
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
}

func executeMetrics(cmd *cobra.Command, args []string) {
	cl := connect()
	ctx, _ := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	for {
		cl.PrintServerState(ctx)
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func connect() *client.Client {
	certificates := readClientCerts()
	cfg := &tls.Config{
		Certificates:     certificates,
		RootCAs:          readCACerts(viper.GetString("ca_cert")),
		VerifyConnection: verifyConnection,
	}
	cred := credentials.NewTLS(cfg)
	cl, err := client.NewClient(viper.GetString("connect"), cred)
	if err != nil {
		panic(err)
	}
	return cl
}

func readClientCerts() []tls.Certificate {
	var certificates []tls.Certificate
	if cert, err := tls.LoadX509KeyPair(viper.GetString("client_cert"), viper.GetString("client_key")); err == nil {
		certificates = append(certificates, cert)
	} else {
		logging.L().Err(err).Msg("Could not read client certificate/key")
	}
	return certificates
}

func verifyConnection(state tls.ConnectionState) error {
	spew.Dump(state.PeerCertificates[0].Subject.CommonName)
	return nil
}

func readCACerts(caCertFiles ...string) *x509.CertPool {
	caCerts := x509.NewCertPool()
	for _, caCertFile := range caCertFiles {
		if caCertData, err := ioutil.ReadFile(caCertFile); err == nil {
			if !caCerts.AppendCertsFromPEM(caCertData) {
				logging.L().Error().Str("filename", caCertFile).Msg("Could not parse CA Certificate from PEM")
			}
		} else {
			logging.L().Err(err).Msg("Could not read CA Certificate")
		}
	}
	return caCerts
}
