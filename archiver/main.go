package main

// TODO: version and tag globals and build tag
// TODO: health checks
// TODO: compression middleware
// TODO: automatic consul support
// TODO: tracing
// TODO: ssl
// TODO: log buffer
import (
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/markbates/pkger"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jeks313/activemq-archiver/archive"
	"github.com/jeks313/activemq-archiver/consumer"
	"github.com/jeks313/activemq-archiver/options"
	"github.com/jeks313/activemq-archiver/server"
	"github.com/rs/zerolog"
)

var opts struct {
	ActiveMQ    ActivemqOpts               `group:"ActiveMQ Options"`
	Service     options.ServiceOptions     `group:"Default Service Options"`
	Application options.ApplicationOptions `group:"Default Application Server Options"`
	// local options here
	Port int `long:"port" env:"PORT" description:"application port" default:"8080"`
}

// ActivemqOpts command line options for activemq
type ActivemqOpts struct {
	Topic    string `long:"topic" env:"TOPIC" description:"topic to archive" required:"true"`
	Hostname string `long:"activemq" env:"ACTIVE_MQ" description:"activemq hostname" default:"localhost:61613"`
	Key      string `long:"key" env:"TOPIC_KEY" description:"key to look for in the document to use to construct archive filename" required:"true"`
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// ensure standard logger is also handled by zerolog
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	stdlog.SetFlags(0)
	stdlog.SetOutput(log)

	_, err := flags.ParseArgs(&opts, os.Args[1:])
	if err != nil {
		log.Error().Err(err).Msg("failed to parse command line arguments")
		os.Exit(1)
	}

	if opts.Application.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// set config from default environment variables / config files
	options.Environment(opts.Application.Environment)

	// define router:
	r := mux.NewRouter()
	r.Use(handlers.CompressHandler)

	// documentation, serves README.md
	f, err := pkger.Open("/README.md")
	if err != nil {
		log.Error().Err(err).Msg("failed to load README.md")
		os.Exit(1)
	}
	server.File(r, "/README.md", server.FormatMarkdown(f))
	server.Redirect(r, "/", "/README.md")

	// logging
	server.Log(r, log)

	// mount default endpoints:
	// debug/profiling
	server.Profiling(r, "/debug/pprof")

	// metrics
	server.Metrics(r, "/metrics")

	// log
	server.LogHandler(r, "/log", 5000)

	// setup health with dependencies.
	server.Health(r, "/status/")

	// Not found if you want to customize
	r.NotFoundHandler = r.NewRoute().HandlerFunc(http.NotFound).GetHandler()

	listen := fmt.Sprintf(":%d", opts.Port)
	srv := &http.Server{
		Handler: r,
		Addr:    listen,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	defer func() {
		signal.Stop(c)
		cancel()
	}()

	a := archive.New()
	q := consumer.Queue{
		Hostname: opts.ActiveMQ.Hostname,
		Topic:    opts.ActiveMQ.Topic,
		Ctx:      ctx,
	}

	log.Info().Str("hostname", opts.ActiveMQ.Hostname).Msg("activemq: connecting")
	err = q.Connect(opts.ActiveMQ.Hostname)
	if err != nil {
		log.Error().Err(err).Str("hostname", opts.ActiveMQ.Hostname).Msg("failed to connect to activemq")
		os.Exit(1)
	}

	log.Info().Str("hostname", opts.ActiveMQ.Hostname).Str("topic", opts.ActiveMQ.Topic).Msg("activemq: subscribing")
	err = q.Subscribe(opts.ActiveMQ.Topic)
	if err != nil {
		log.Error().Err(err).Str("topic", opts.ActiveMQ.Topic).Msg("unable to subscribe to topic")
		os.Exit(1)
	}

	log.Info().Str("hostname", opts.ActiveMQ.Hostname).Str("topic", opts.ActiveMQ.Topic).Msg("activemq: consuming")
	go q.Consume(a)

	go func() {
		select {
		case <-c:
			cancel()
			log.Info().Msg("shutting down http listener ...")
			srv.Shutdown(ctx)
		case <-ctx.Done():
		}
	}()

	log.Info().Int("port", opts.Port).Msg("starting listening server")

	err = srv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("failed to start http server")
		os.Exit(1)
	}

	log.Info().Msg("stopped")
}
