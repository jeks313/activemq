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
	"github.com/jeks313/activemq-archiver/internal/archive"
	"github.com/jeks313/activemq-archiver/internal/consumer"
	"github.com/jeks313/activemq-archiver/internal/content"
	"github.com/jeks313/activemq-archiver/pkg/options"
	"github.com/jeks313/activemq-archiver/pkg/server"
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
	Topic       string `long:"topic" env:"TOPIC" description:"topic to archive" required:"true"`
	Hostname    string `long:"activemq" env:"ACTIVE_MQ" description:"activemq hostname" default:"localhost:61613"`
	ArchivePath string `long:"archive-path" env:"ARCHIVE_PATH" default:"/var/lib/activemq-archive" description:"base directory to write archive files"`
	Key         string `long:"key" env:"TOPIC_KEY" description:"key to look for in the document to use to construct archive filename" required:"true"`
	MaxSize     int    `long:"max-size" env:"MAX_ARCHIVE_SIZE" description:"maximum archive size, defaults to 32M for athena usage in S3" default:"33554432"`
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

	if opts.Application.Version {
		options.LogVersion()
		os.Exit(0)
	}

	if _, err := os.Stat(opts.ActiveMQ.ArchivePath); os.IsNotExist(err) {
		log.Error().Err(err).Str("archive_path", opts.ActiveMQ.ArchivePath).Msg("output path does not exist")
		os.Exit(1)
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

	a := archive.New(opts.ActiveMQ.MaxSize)
	a.Path = opts.ActiveMQ.ArchivePath

	q := consumer.New()
	q.Hostname = opts.ActiveMQ.Hostname
	q.Topic = opts.ActiveMQ.Topic
	q.Key = opts.ActiveMQ.Key
	q.Ctx = ctx
	q.ContentTypeHeader = "X-Content-Type"
	q.Handlers["zip;json"] = content.HandlerFromZipJSON
	q.Handlers["b64;zip;json"] = content.HandlerFromBase64ZipJSON

	// main activemq consumer loop, reconnects with 2 sec delay
	go func() {
		var err error
		for {
			log.Info().Str("hostname", opts.ActiveMQ.Hostname).Msg("activemq: connecting")
			err = q.Connect(opts.ActiveMQ.Hostname)
			if err != nil {
				log.Error().Err(err).Str("hostname", opts.ActiveMQ.Hostname).Msg("failed to connect to activemq")
				continue
			}

			log.Info().Str("hostname", opts.ActiveMQ.Hostname).Str("topic", opts.ActiveMQ.Topic).Msg("activemq: subscribing")
			err = q.Subscribe(opts.ActiveMQ.Topic)
			if err != nil {
				log.Error().Err(err).Str("topic", opts.ActiveMQ.Topic).Msg("unable to subscribe to topic")
				continue
			}

			log.Info().Str("hostname", opts.ActiveMQ.Hostname).Str("topic", opts.ActiveMQ.Topic).Msg("activemq: consuming")
			err = q.Consume(a)

			if err != nil {
				log.Error().Err(err).Msg("consumer ended")
			}

			log.Info().Msg("waiting to re-connect...")
			time.Sleep(2 * time.Second)
		}
	}()

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
