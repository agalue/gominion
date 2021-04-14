package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/broker"
	"github.com/agalue/gominion/collectors"
	"github.com/agalue/gominion/detectors"
	"github.com/agalue/gominion/log"
	"github.com/agalue/gominion/monitors"

	_ "github.com/agalue/gominion/rpc"  // Load all RPC modules
	_ "github.com/agalue/gominion/sink" // Load all Sink modules

	homedir "github.com/mitchellh/go-homedir"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// cfgFile is the path to the Minion configuration file
	cfgFile string

	// listeners a list of Sink API listeners
	listeners = []string{}

	// minionConfig is the Minion configuration with defaults
	minionConfig = &api.MinionConfig{
		BrokerType: "grpc",
		Location:   "Local",
		BrokerURL:  "localhost:8990",
		TrapPort:   1162,
		SyslogPort: 1514,
		LogLevel:   "debug",
	}

	// rootCmd represents the base command that starts the Minion's gRPC client
	rootCmd = &cobra.Command{
		Use:     "gominion",
		Short:   "An implementation of OpenNMS Minion in Go",
		Version: "0.1.0-alpha1",
		Run:     rootHandler,
		Args:    cobra.NoArgs,
	}
)

// Execute prepares the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Initialize Configuration
	cobra.OnInitialize(initConfig)

	// Initialize Flags
	hostname, _ := os.Hostname()
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ~/.gominion.yaml)")
	rootCmd.Flags().StringVarP(&minionConfig.ID, "id", "i", hostname, "Minion ID")
	rootCmd.Flags().StringVarP(&minionConfig.Location, "location", "l", minionConfig.Location, "Minion Location")
	rootCmd.Flags().StringVarP(&minionConfig.BrokerType, "brokerType", "b", minionConfig.BrokerURL, "Broker Type, either grpc or kafka")
	rootCmd.Flags().StringVarP(&minionConfig.BrokerURL, "brokerUrl", "u", minionConfig.BrokerType, "Broker URL")
	rootCmd.Flags().IntVarP(&minionConfig.TrapPort, "trapPort", "t", minionConfig.TrapPort, "SNMP Trap port")
	rootCmd.Flags().IntVarP(&minionConfig.SyslogPort, "syslogPort", "s", minionConfig.SyslogPort, "Syslog port")
	rootCmd.Flags().IntVarP(&minionConfig.StatsPort, "statsPort", "S", minionConfig.StatsPort, "HTTP Prometheus exporter statistics port")
	rootCmd.Flags().StringArrayVarP(&listeners, "listener", "L", nil, "Flow/Telemetry listeners\ne.x. -L Graphite,2003,ForwardParser -L NXOS,5000,NxosGrpcParser")
	rootCmd.Flags().StringVarP(&minionConfig.LogLevel, "logLevel", "x", minionConfig.LogLevel, "Logging level")

	// Initialize Flag Binding
	viper.BindPFlags(rootCmd.Flags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigType("yaml")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name "gominion" (without extension).
		if home, err := homedir.Dir(); err != nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
		viper.SetConfigName(".gominion")
	}

	viper.SetEnvPrefix("GOMINION")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file:", viper.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Warnf("Cannot read configuration file: %v", err)
		}
	}
	if err := viper.Unmarshal(minionConfig); err != nil {
		log.Warnf("Cannot parse configuration file: %v", err)
	}
}

func displayRegisteredModules() {
	for _, m := range api.GetAllRPCModules() {
		log.Debugf("Registered RPC module %s", m.GetID())
	}
	for _, m := range api.GetAllSinkModules() {
		log.Debugf("Registered Sink module %s", m.GetID())
	}
	for _, m := range collectors.GetAllCollectors() {
		log.Debugf("Registered collector module %s", m.GetID())
	}
	for _, m := range detectors.GetAllDetectors() {
		log.Debugf("Registered detector module %s", m.GetID())
	}
	for _, m := range monitors.GetAllMonitors() {
		log.Debugf("Registered poller module %s", m.GetID())
	}
}

func rootHandler(cmd *cobra.Command, args []string) {
	log.InitLogger(minionConfig.LogLevel)
	// Validate Configuration
	if err := minionConfig.IsValid(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	if err := minionConfig.ParseListeners(listeners); err != nil {
		log.Fatalf("Invalid listener configuration: %v", err)
	}
	// Display loaded modules
	displayRegisteredModules()
	// Start statistics server
	if minionConfig.StatsPort > 0 {
		go func() {
			log.Infof("Starting Prometheus Metrics server on port %d", minionConfig.StatsPort)
			http.Handle("/", promhttp.Handler())
			err := http.ListenAndServe(fmt.Sprintf(":%d", minionConfig.StatsPort), nil)
			if err != nil {
				log.Fatalf("Cannot start prometheus HTTP server: %v", err)
			}
		}()
	}
	// Start client
	client := broker.GetBroker(minionConfig)
	if client == nil {
		log.Fatalf("Cannot find broker implementation for %s", minionConfig.BrokerType)
	}
	log.Infof("Starting OpenNMS Minion...\n%s", minionConfig.String())
	metrics := api.NewMetrics()
	if minionConfig.StatsPort > 0 {
		metrics.Register()
	}
	if err := client.Start(minionConfig, metrics); err != nil {
		log.Fatalf("Cannot connect via %s: %v", minionConfig.BrokerType, err)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	client.Stop()
}
